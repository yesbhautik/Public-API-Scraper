# PublicAPI-Scraper
A tool to find and verify exposed API keys in public GitHub repositories. This tool helps identify potentially leaked API keys and allows for responsible disclosure to API providers.

## Features
- Web-based user interface for easy interaction
- Real-time search progress and results
- Multi-architecture support (AMD64 and ARM64)
- Live verification of found API keys
- WebSocket-based communication for real-time updates
- Docker support for easy deployment

## Prerequisites
- Docker installed on your system
- [GitHub Personal Access Token](https://github.com/settings/personal-access-tokens/new) with repo and read:packages scopes (Public Repositories READ permissions works fine)

## Quick Start
1. Create a .env file with your GitHub token:
> If you pass token in command line, you don't need to create .env file
```
GITHUB_TOKEN=your_github_token_here
```



2. Run using Docker:
```
docker run -p 3000:3000 -e GITHUB_TOKEN=your_github_token yesbhautik/public-api-scraper:latest
```

Or using docker-compose:
```
docker-compose up -d
```

3. Open your browser and visit `http://localhost:3000`

## Usage
1. Enter the required information in the web interface:
- Model: The AI model identifier (e.g., `gpt-4`, `deepseek-ai/deepseek-coder`)
- Endpoint: The API endpoint URL (e.g., `https://api.openai.com/v1/chat/completions`)
- Search Keyword: The keyword to search for (e.g., `openai_api`, `anthropic_api_key`)
- Click "Start Search" to begin the process


2. Monitor the logs section for real-time search progress
3. View working API keys in the results section


## Building from Source
1. Clone the repository:
```
git clone https://github.com/yesbhautik/PublicAPI-Scraper.git
cd PublicAPI-Scraper
```


2. Build the Docker image:
```
docker build -t public-api-scraper .
```

3. Run using Docker:
```
docker run -p 3000:3000 -e GITHUB_TOKEN=your_github_token public-api-scraper
```
    
## Development Setup
1. Install dependencies:
```
go mod download
```

2. Run the development server:
```
go run main.go
```

3. Open your browser and visit `http://localhost:3000`

## Project Structure
```
public-api-scraper/
├── main.go                 # Main application entry point
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
├── templates/
│   └── index.html        # Web UI template
├── internal/
│   ├── github/           # GitHub search functionality
│   │   └── search.go
│   └── verifier/         # API key verification
│       └── api_verifier.go
└── .env                  # Environment configuration
```

## Configuration
Environment Variables:
`GITHUB_TOKEN:` Your [GitHub Personal Access Token](https://github.com/settings/personal-access-tokens/new)
`PORT:` Server port (default: 3000)



## Security Considerations
- Store your GitHub token securely
- Use this tool responsibly for security research


## Report found API keys to their respective owners
- Do not use or share found API keys

## Contributing
1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request


## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer
This tool is meant for security research and responsible disclosure. Do not use found API keys or use this tool for malicious purposes.

## Support
For support, please open an issue in the GitHub repository.