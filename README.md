# gitignore.lol 🚀

> For devs who hate commit noise

No redirects. No ads. Just clean, GitHub-powered templates.

## Overview

gitignore.lol is a modern, fast, and simple service for generating `.gitignore` files for your projects. Inspired by the classic gitignore.io, but built with modern technologies and a focus on simplicity and performance.

## Features

- 🔗 **No Redirects** - No weird rebranding and redirects. The base API URL stays the same.
- 🚀 **Zero Fuss** - Quick and simple .gitignore generation powered by GitHub's official templates.
- 💻 **Web or CLI** - Generate templates through the web interface or use the REST API - whatever fits your workflow best.
- ⚡ **Fast & Reliable** - Built with Go for maximum performance and reliability.
- 🔒 **Rate Limited** - Protected against abuse with sensible rate limits.
- 🌐 **CORS Enabled** - Ready for cross-origin requests.

## Quick Start

### Using the API

Get a list of available templates:

```bash
curl https://gitignore.lol/api/list
```

Generate a .gitignore file for Go and Node.js:

```bash
curl https://gitignore.lol/api/go,node
```

### Building from Source

Requirements:

- Go 1.21.6 or later
- Git

```bash
# Clone the repository
git clone https://github.com/valerius21/gitignore.lol.git
cd gitignore.lol

# Install dependencies
go mod download

# Build
go build -o gitignore-server ./cmd/gitignore_server.go

# Run
./gitignore-server
```

## API Documentation

The API is documented using OpenAPI/Swagger. You can access the documentation at:

- [Swagger UI](https://gitignore.lol/swagger/index.html)
- [OpenAPI Spec](https://gitignore.lol/swagger/doc.json)

### Endpoints

- `GET /api/list` - Get a list of all available templates
- `GET /api/{templateList}` - Get combined .gitignore file for specified templates
  - Example: `/api/go,node,python`

## Development

### Prerequisites

- Go 1.21.6 or later
- Docker (optional)
- Git

### Setup

1. Clone the repository

```bash
git clone https://github.com/valerius21/gitignore.lol.git
cd gitignore.lol
```

2. Install dependencies

```bash
go mod download
```

3. Run tests

```bash
go test ./...
```

4. Generate Swagger documentation

```bash
./scripts/generate-swagger.sh
```

### Environment Variables

- `PORT` - Server port (default: 3000)
- `LOG_LEVEL` - Logging level (default: warn)

### Running the server

```bash
$ go run ./cmd/main.go --help
Usage: main [flags]

Flags:
  -h, --help                                              Show context-sensitive help.
      --port=4444                                         Port the server listens on.
      --repo="https://github.com/github/gitignore.git"    Gitignore repository where the .gitignore files are versioned.
      --clone-path="./store"                              Location of the locally stored gitignore repository
      --fetch-interval=300                                Interval (seconds) in which the linked repository gets updated
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

- Project URL: [https://github.com/valerius21/gitignore.lol](https://github.com/valerius21/gitignore.lol)
