#!/bin/bash

# 📊 Run All Benchmarks Script
# Runs benchmarks for all 6 projects to measure performance

set -e  # Exit on error

echo "=================================="
echo "Running All Project Benchmarks"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
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

# Benchmark time (adjust if needed)
BENCH_TIME="3s"

# Run benchmarks for each project
for project in "${PROJECTS[@]}"; do
    echo "======================================"
    echo -e "${BLUE}Benchmarking: $project${NC}"
    echo "======================================"
    
    if [ ! -d "$project/final" ]; then
        echo -e "${YELLOW}⚠️  Warning: $project/final not found, skipping...${NC}"
        echo ""
        continue
    fi
    
    cd "$project/final"
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        echo -e "${YELLOW}⚠️  Warning: go.mod not found. Run ./setup.sh first!${NC}"
        cd ../..
        echo ""
        continue
    fi
    
    # Run benchmarks
    echo "Running: go test -bench=. -benchmem -benchtime=$BENCH_TIME"
    echo ""
    
    go test -bench=. -benchmem -benchtime=$BENCH_TIME 2>/dev/null || echo -e "${YELLOW}No benchmarks found or benchmark failed${NC}"
    
    cd ../..
    echo ""
done

echo "======================================"
echo -e "${GREEN}✅ Benchmark suite complete!${NC}"
echo "======================================"
echo ""
echo "Performance targets:"
echo "  • rate-limiter:      500k+ ops/sec with sharding"
echo "  • job-queue:         10k+ jobs/sec"
echo "  • cache:             100M+ ops/sec with 256 shards"
echo "  • web-crawler:       200-500 pages/min (rate limited)"
echo "  • connection-pool:   50k+ ops/sec"
echo "  • pub-sub:           100k+ msgs/sec"
echo ""
echo "💡 Tip: Compare your results with the targets above!"
