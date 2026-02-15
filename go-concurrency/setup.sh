#!/bin/bash

# 🚀 Go Concurrency Projects Setup Script
# This script initializes all 6 projects with Go modules

set -e  # Exit on error

echo "=================================="
echo "Go Concurrency Projects Setup"
echo "=================================="
echo ""

# Check Go version
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed!"
    echo "Please install Go 1.19+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Found $GO_VERSION"
echo ""

# Project list
PROJECTS=(
    "rate-limiter"
    "job-queue"
    "cache"
    "web-crawler"
    "connection-pool"
    "pub-sub"
)

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/projects"

echo "Initializing projects..."
echo ""

# Initialize each project
for project in "${PROJECTS[@]}"; do
    echo "📦 Setting up $project..."
    
    if [ ! -d "$project/final" ]; then
        echo "⚠️  Warning: $project/final directory not found, skipping..."
        continue
    fi
    
    cd "$project/final"
    
    # Initialize go module if go.mod doesn't exist
    if [ ! -f "go.mod" ]; then
        go mod init "github.com/yourusername/go-concurrency/$project" > /dev/null 2>&1
        echo "   ✓ Created go.mod"
    else
        echo "   ✓ go.mod exists"
    fi
    
    # Special handling for web-crawler (needs external dependency)
    if [ "$project" = "web-crawler" ]; then
        echo "   → Downloading golang.org/x/net/html..."
        go get golang.org/x/net/html > /dev/null 2>&1
    fi
    
    # Tidy up dependencies
    go mod tidy > /dev/null 2>&1
    echo "   ✓ Dependencies resolved"
    
    cd ../..
    echo ""
done

echo "=================================="
echo "✅ Setup Complete!"
echo "=================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Run tests for a specific project:"
echo "   cd projects/rate-limiter/final"
echo "   go test -v"
echo ""
echo "2. Run all tests:"
echo "   ./run_all_tests.sh"
echo ""
echo "3. Run benchmarks:"
echo "   cd projects/rate-limiter/final"
echo "   go test -bench=. -benchmem"
echo ""
echo "4. Read the getting started guide:"
echo "   cat GETTING_STARTED.md"
echo ""
echo "Happy learning! 🎓"
