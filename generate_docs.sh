#!/bin/zsh

echo "Generating Swagger documentation..."

# Check if swag is installed
if ! command -v $HOME/go/bin/swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate documentation
$HOME/go/bin/swag init -g pkg/server/api.go --parseDependency --parseInternal

# Check if generation was successful
if [ $? -eq 0 ]; then
    echo "✅ Documentation generated successfully!"
    echo "📚 View the documentation at http://localhost:4444/swagger/index.html when the server is running"
else
    echo "❌ Failed to generate documentation"
    exit 1
fi 