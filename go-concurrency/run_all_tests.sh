#!/bin/bash

# 🧪 Run All Tests Script
# Runs tests for all 6 projects with race detector

set -e  # Exit on error

echo "=================================="
echo "Running All Project Tests"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Run tests for each project
for project in "${PROJECTS[@]}"; do
    echo "======================================"
    echo "Testing: $project"
    echo "======================================"
    
    if [ ! -d "$project/final" ]; then
        echo -e "${YELLOW}⚠️  Warning: $project/final not found, skipping...${NC}"
        echo ""
        continue
    fi
    
    cd "$project/final"
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}❌ Error: go.mod not found. Run ./setup.sh first!${NC}"
        cd ../..
        continue
    fi
    
    # Run tests with race detector
    echo "Running: go test -race -v"
    echo ""
    
    if go test -race -v; then
        echo ""
        echo -e "${GREEN}✅ $project: PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo ""
        echo -e "${RED}❌ $project: FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    cd ../..
    echo ""
done

# Summary
echo "======================================"
echo "Test Summary"
echo "======================================"
echo "Total Projects: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
else
    echo "Failed: 0"
fi

echo ""
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check output above for details.${NC}"
    exit 1
fi
