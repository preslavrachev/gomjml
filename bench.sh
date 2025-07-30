#!/bin/bash

# MJML Benchmark Runner with Progress and Markdown Export
# Usage: ./bench.sh [--markdown|-m]

MARKDOWN=false
if [[ "$1" == "--markdown" || "$1" == "-m" ]]; then
    MARKDOWN=true
fi

echo "🚀 Starting MJML benchmarks..."
echo "📊 Running comprehensive performance tests..."

# Process benchmark results
go test ./mjml -bench=. -benchmem -run='^$' 2>/dev/null | awk -v markdown="$MARKDOWN" '
/^Benchmark/ {
    if (/ns\/op.*B\/op.*allocs\/op/) {
        # Has memory data
        time_ns = $3
        time_ms = time_ns / 1000000
        bytes = $5
        mb = bytes / (1024 * 1024)
        allocs = $7
        allocs_k = allocs / 1000
        
        if (markdown == "true") {
            printf "| %-33s | %6.2fms | %6.2fMB | %8.1fK |\n", $1, time_ms, mb, allocs_k
        } else {
            printf "%-35s %8.2fms %8.2fMB %10.1fK\n", $1, time_ms, mb, allocs_k
        }
    } else if (/ns\/op/) {
        # Time only
        time_ns = $3
        time_ms = time_ns / 1000000
        
        if (markdown == "true") {
            printf "| %-33s | %6.2fms | %8s | %8s |\n", $1, time_ms, "-", "-"
        } else {
            printf "%-35s %8.2fms %8s %10s\n", $1, time_ms, "-", "-"
        }
    }
    next
}
BEGIN {
    if (markdown == "true") {
        printf "| %-33s | %8s | %8s | %8s |\n", "Benchmark", "Time", "Memory", "Allocs"
        printf "|%-34s|%9s|%9s|%9s|\n", ":---------------------------------", ":-------:", ":-------:", ":-------:"
    } else {
        printf "%-35s %8s %8s %10s\n", "Benchmark", "Time", "Memory", "Allocs"
        printf "%-35s %8s %8s %10s\n", "---------", "----", "------", "------"
    }
}' | if [[ "$MARKDOWN" == "false" ]]; then column -t; else cat; fi


if [[ "$MARKDOWN" == "true" ]]; then
    echo ""
    echo "📋 Markdown table generated! Copy the output above to use in documentation."
else
    echo ""
    echo "💡 Tip: Use './bench.sh --markdown' to generate markdown table format"
fi