package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"api-key-finder/internal/github"
	"api-key-finder/internal/verifier"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SearchRequest struct {
	Model    string `json:"model"`
	Endpoint string `json:"endpoint"`
	Keyword  string `json:"keyword"`
}

type WSMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	API     string `json:"api,omitempty"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Serve static files
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/", fs)

	// WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	log.Printf("Server starting at :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var searchReq SearchRequest
		if err := json.Unmarshal(message, &searchReq); err != nil {
			log.Println(err)
			continue
		}

		// Start the search process
		go processSearch(conn, searchReq)
	}
}

func processSearch(conn *websocket.Conn, req SearchRequest) {
	log.Printf("Starting search with keyword: %s", req.Keyword)

	// Send initial log message
	sendWSMessage(conn, WSMessage{
		Type:    "log",
		Message: "Starting search process...",
	})

	// Initialize GitHub searcher with token
	githubToken := os.Getenv("GITHUB_TOKEN")
	log.Printf("Token length: %d, First 10 chars: %s", len(githubToken), githubToken[:10])
	if githubToken == "" {
		log.Println("GitHub token not found")
		sendWSMessage(conn, WSMessage{
			Type:    "log",
			Message: "Error: GITHUB_TOKEN environment variable not set",
		})
		return
	}

	log.Printf("Using GitHub token: %s...", githubToken[:10])
	searcher := github.NewGithubSearcher(githubToken)

	// Search for API keys
	sendWSMessage(conn, WSMessage{
		Type:    "log",
		Message: "Searching GitHub for potential API keys...",
	})

	results, err := searcher.Search(req.Keyword)
	if err != nil {
		log.Printf("Search error: %v", err)
		sendWSMessage(conn, WSMessage{
			Type:    "log",
			Message: "Error searching GitHub: " + err.Error(),
		})
		return
	}

	log.Printf("Found %d results", len(results))

	// Initialize API verifier
	apiVerifier := verifier.NewAPIVerifier(req.Endpoint, req.Model)

	// Verify each API key
	for _, result := range results {
		sendWSMessage(conn, WSMessage{
			Type:    "log",
			Message: fmt.Sprintf("Verifying API key from %s...", result.RepoName),
		})

		if apiVerifier.Verify(result.APIKey) {
			sendWSMessage(conn, WSMessage{
				Type:    "result",
				API:     result.APIKey,
				Message: fmt.Sprintf("Working API key found in %s (%s)", result.RepoName, result.FilePath),
			})
		}
	}

	sendWSMessage(conn, WSMessage{
		Type:    "log",
		Message: "Search process completed",
	})
}

func sendWSMessage(conn *websocket.Conn, msg WSMessage) {
	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Error sending message:", err)
	}
}
