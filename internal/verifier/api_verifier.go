package verifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type APIVerifier struct {
	Endpoint string
	Model    string
}

func NewAPIVerifier(endpoint, model string) *APIVerifier {
	return &APIVerifier{
		Endpoint: endpoint,
		Model:    model,
	}
}

func (v *APIVerifier) Verify(apiKey string) bool {
	jsonData := map[string]interface{}{
		"model":       v.Model,
		"max_tokens":  20,
		"temperature": 0.6,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Hello, test message",
			},
		},
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", v.Endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return false
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
