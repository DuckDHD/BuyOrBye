#!/bin/bash

# Production deployment script

set -e

echo "🚀 Deploying Should I Get It..."

# Build the application
echo "🔧 Building application..."
make build

# Run tests
echo "🧪 Running tests..."
make test

# Build Docker image
echo "🐳 Building Docker image..."
docker build -t should-i-get-it:latest .

echo "✅ Build complete!"
echo "Docker image: should-i-get-it:latest"