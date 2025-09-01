#!/bin/bash

# Development setup script

set -e

echo "🚀 Setting up Should I Get It development environment..."

# Check Go version
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24+"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3 | cut -c3-)
echo "✅ Go version: $GO_VERSION"

# Install templ if not already installed
if ! command -v templ &> /dev/null; then
    echo "📦 Installing templ..."
    go install github.com/a-h/templ/cmd/templ@latest
fi

# Install air for hot reload if not already installed
if ! command -v air &> /dev/null; then
    echo "📦 Installing air..."
    go install github.com/air-verse/air@latest
fi

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "⚠️  Please update .env with your actual configuration"
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Install Node dependencies
if [ -f package.json ]; then
    echo "📦 Installing Node dependencies..."
    npm install
fi

# Generate templ templates
echo "🔧 Generating templ templates..."
templ generate

# Build CSS
if [ -f package.json ]; then
    echo "🎨 Building CSS..."
    npm run build-css-prod
fi

# Create directories if they don't exist
mkdir -p {bin,tmp,static/css,static/js,static/images}

echo "✅ Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Update your .env file with database credentials"
echo "2. Run 'make docker-up' to start PostgreSQL and Redis"
echo "3. Run 'make migrate-up' to create database tables"
echo "4. Run 'make dev' to start the development server"