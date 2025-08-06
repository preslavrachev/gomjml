#!/bin/bash

# Benchmark gomjml vs mrml on Austin template
TEMPLATE="mjml/testdata/austin-layout-from-mjml-io.mjml"
ITERATIONS=50
MARKDOWN_OUTPUT=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --markdown|-md)
            MARKDOWN_OUTPUT=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--markdown|--md] [--help|-h]"
            echo "  --markdown, -md    Output results table in markdown format"
            echo "  --help, -h         Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "🏁 Benchmarking gomjml vs mrml vs mjml on Austin template"
echo "📄 Template: $TEMPLATE"
echo "🔄 Iterations: $ITERATIONS"
if [[ "$MARKDOWN_OUTPUT" == "true" ]]; then
    echo "📝 Output format: Markdown"
fi
echo ""

# Build gomjml
echo "🔨 Building gomjml..."
go build -o bin/gomjml ./cmd/gomjml

echo "⏱️  Running benchmarks..."
echo ""

# PASS 1: Pure timing measurements (no monitoring overhead)
echo "🏃 PASS 1: Timing measurements"
echo ""

# Benchmark gomjml
echo "🐹 Testing gomjml..."
gomjml_start=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
for i in $(seq 1 $ITERATIONS); do
    ./bin/gomjml compile "$TEMPLATE" -o /tmp/gomjml-output.html >/dev/null 2>&1
done
gomjml_end=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
gomjml_time=$(echo "($gomjml_end - $gomjml_start) * 1000" | bc -l | cut -d. -f1)

# Benchmark mrml
echo "🦀 Testing mrml..."
mrml_start=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
for i in $(seq 1 $ITERATIONS); do
    mrml render "$TEMPLATE" -o /tmp/mrml-output.html >/dev/null 2>&1
done
mrml_end=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
mrml_time=$(echo "($mrml_end - $mrml_start) * 1000" | bc -l | cut -d. -f1)

# Benchmark mjml (JS)
echo "🟨 Testing mjml (JS)..."
mjml_start=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
for i in $(seq 1 $ITERATIONS); do
    mjml -r "$TEMPLATE" -o /tmp/mjml-output.html >/dev/null 2>&1
done
mjml_end=$(perl -MTime::HiRes -e 'print Time::HiRes::time()')
mjml_time=$(echo "($mjml_end - $mjml_start) * 1000" | bc -l | cut -d. -f1)

echo ""
echo "🔍 PASS 2: Resource monitoring"
echo ""

# Function to monitor process resources
monitor_resources() {
    local cmd="$1"
    local tool_name="$2"
    
    # Start the process in background
    $cmd >/dev/null 2>&1 &
    local pid=$!
    
    local max_memory=0
    local cpu_samples=0
    local cpu_total=0
    
    # Monitor while process is running
    while kill -0 "$pid" 2>/dev/null; do
        # Get memory usage in MB
        if [[ "$OSTYPE" == "darwin"* ]]; then
            local memory=$(ps -o rss= -p "$pid" 2>/dev/null | awk '{print int($1/1024)}')
            local cpu=$(ps -o %cpu= -p "$pid" 2>/dev/null | awk '{print $1}')
        else
            local memory=$(ps -o rss= -p "$pid" 2>/dev/null | awk '{print int($1/1024)}')
            local cpu=$(ps -o %cpu= -p "$pid" 2>/dev/null | awk '{print $1}')
        fi
        
        if [[ -n "$memory" && "$memory" -gt "$max_memory" ]]; then
            max_memory=$memory
        fi
        
        if [[ -n "$cpu" && "$cpu" != "" ]]; then
            cpu_total=$(echo "$cpu_total + $cpu" | bc -l 2>/dev/null || echo "$cpu_total")
            ((cpu_samples++))
        fi
        
        sleep 0.01  # Sample every 10ms
    done
    
    wait "$pid"
    
    # Calculate average CPU
    local avg_cpu=0
    if [[ $cpu_samples -gt 0 ]]; then
        avg_cpu=$(echo "scale=1; $cpu_total / $cpu_samples" | bc -l 2>/dev/null || echo "0")
    fi
    
    echo "$max_memory $avg_cpu"
}

# Monitor gomjml resources (50 iterations)
echo "🐹 Monitoring gomjml resources..."
gomjml_max_memory=0
gomjml_cpu_total=0
for i in $(seq 1 $ITERATIONS); do
    gomjml_resources=$(monitor_resources "./bin/gomjml compile '$TEMPLATE' -o /tmp/gomjml-monitor-$i.html" "gomjml")
    memory=$(echo $gomjml_resources | cut -d' ' -f1)
    cpu=$(echo $gomjml_resources | cut -d' ' -f2)
    
    if [[ $memory -gt $gomjml_max_memory ]]; then
        gomjml_max_memory=$memory
    fi
    gomjml_cpu_total=$(echo "$gomjml_cpu_total + $cpu" | bc -l 2>/dev/null || echo "$gomjml_cpu_total")
done
gomjml_avg_cpu=$(echo "scale=1; $gomjml_cpu_total / $ITERATIONS" | bc -l 2>/dev/null || echo "0")

# Monitor mrml resources (50 iterations)
echo "🦀 Monitoring mrml resources..."
mrml_max_memory=0
mrml_cpu_total=0
for i in $(seq 1 $ITERATIONS); do
    mrml_resources=$(monitor_resources "mrml render '$TEMPLATE' -o /tmp/mrml-monitor-$i.html" "mrml")
    memory=$(echo $mrml_resources | cut -d' ' -f1)
    cpu=$(echo $mrml_resources | cut -d' ' -f2)
    
    if [[ $memory -gt $mrml_max_memory ]]; then
        mrml_max_memory=$memory
    fi
    mrml_cpu_total=$(echo "$mrml_cpu_total + $cpu" | bc -l 2>/dev/null || echo "$mrml_cpu_total")
done
mrml_avg_cpu=$(echo "scale=1; $mrml_cpu_total / $ITERATIONS" | bc -l 2>/dev/null || echo "0")

# Monitor mjml resources (50 iterations)
echo "🟨 Monitoring mjml resources..."
mjml_max_memory=0
mjml_cpu_total=0
for i in $(seq 1 $ITERATIONS); do
    mjml_resources=$(monitor_resources "mjml -r '$TEMPLATE' -o /tmp/mjml-monitor-$i.html" "mjml")
    memory=$(echo $mjml_resources | cut -d' ' -f1)
    cpu=$(echo $mjml_resources | cut -d' ' -f2)
    
    if [[ $memory -gt $mjml_max_memory ]]; then
        mjml_max_memory=$memory
    fi
    mjml_cpu_total=$(echo "$mjml_cpu_total + $cpu" | bc -l 2>/dev/null || echo "$mjml_cpu_total")
done
mjml_avg_cpu=$(echo "scale=1; $mjml_cpu_total / $ITERATIONS" | bc -l 2>/dev/null || echo "0")

# Calculate results
gomjml_avg=$(( gomjml_time / ITERATIONS ))
mrml_avg=$(( mrml_time / ITERATIONS ))
mjml_avg=$(( mjml_time / ITERATIONS ))

echo ""
echo "📊 Results:"

if [[ "$MARKDOWN_OUTPUT" == "true" ]]; then
    # Markdown table format
    echo ""
    printf "| %-11s | %-19s | %-11s | %-11s | %-11s |\n" "Tool" "${ITERATIONS}x Total (ms)" "Avg (ms)" "Max RAM (MB)" "Avg CPU (%)"
    printf "|%s|%s|%s|%s|%s|\n" "-------------" "---------------------" "-------------" "-------------" "-------------"
    printf "| %-11s | %19d | %11d | %11d | %11s |\n" "gomjml" $gomjml_time $gomjml_avg $gomjml_max_memory $gomjml_avg_cpu
    printf "| %-11s | %19d | %11d | %11d | %11s |\n" "mrml" $mrml_time $mrml_avg $mrml_max_memory $mrml_avg_cpu
    printf "| %-11s | %19d | %11d | %11d | %11s |\n" "mjml (JS)" $mjml_time $mjml_avg $mjml_max_memory $mjml_avg_cpu
    echo ""
else
    # ASCII table format
    echo "┌─────────────┬─────────────────────┬─────────────┬─────────────┬─────────────┐"
    printf "│ %-11s │ %-19s │ %-11s │ %-11s │ %-11s │\n" "Tool" "${ITERATIONS}x Total (ms)" "Avg (ms)" "Max RAM (MB)" "Avg CPU (%)"
    echo "├─────────────┼─────────────────────┼─────────────┼─────────────┼─────────────┤"
    printf "│ %-11s │ %19d │ %11d │ %11d │ %11s │\n" "gomjml" $gomjml_time $gomjml_avg $gomjml_max_memory $gomjml_avg_cpu
    printf "│ %-11s │ %19d │ %11d │ %11d │ %11s │\n" "mrml" $mrml_time $mrml_avg $mrml_max_memory $mrml_avg_cpu
    printf "│ %-11s │ %19d │ %11d │ %11d │ %11s │\n" "mjml (JS)" $mjml_time $mjml_avg $mjml_max_memory $mjml_avg_cpu
    echo "└─────────────┴─────────────────────┴─────────────┴─────────────┴─────────────┘"
fi

# Find fastest tool
fastest="gomjml"
fastest_time=$gomjml_avg
if [ $mrml_avg -lt $fastest_time ]; then
    fastest="mrml"
    fastest_time=$mrml_avg
fi
if [ $mjml_avg -lt $fastest_time ]; then
    fastest="mjml"
    fastest_time=$mjml_avg
fi

echo ""
echo "🏆 Fastest: $fastest (${fastest_time}ms avg)"

# Performance comparisons
echo ""
echo "📈 Performance comparisons:"
if [ $fastest = "gomjml" ]; then
    mrml_diff=$(( mrml_avg * 100 / gomjml_avg - 100 ))
    mjml_diff=$(( mjml_avg * 100 / gomjml_avg - 100 ))
    echo "  • mrml is ${mrml_diff}% slower than gomjml"
    echo "  • mjml is ${mjml_diff}% slower than gomjml"
elif [ $fastest = "mrml" ]; then
    gomjml_diff=$(( gomjml_avg * 100 / mrml_avg - 100 ))
    mjml_diff=$(( mjml_avg * 100 / mrml_avg - 100 ))
    echo "  • gomjml is ${gomjml_diff}% slower than mrml"
    echo "  • mjml is ${mjml_diff}% slower than mrml"
else
    gomjml_diff=$(( gomjml_avg * 100 / mjml_avg - 100 ))
    mrml_diff=$(( mrml_avg * 100 / mjml_avg - 100 ))
    echo "  • gomjml is ${gomjml_diff}% slower than mjml"
    echo "  • mrml is ${mrml_diff}% slower than mjml"
fi

# Clean up
rm -f /tmp/gomjml-output.html /tmp/mrml-output.html /tmp/mjml-output.html
rm -f /tmp/gomjml-monitor-*.html /tmp/mrml-monitor-*.html /tmp/mjml-monitor-*.html