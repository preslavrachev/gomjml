# Performance Benchmarks

This document provides comprehensive performance benchmarks for gomjml compared to other MJML implementations.

## Overview

The benchmark suite compares three MJML compilation tools using a realistic email template (Austin layout from mjml.io):

- **gomjml** - This native Go implementation
- **mrml** - Rust implementation (reference)
- **mjml (JS)** - The original JavaScript implementation

## Test Methodology

The benchmarking script (`bench-austin.sh`) uses a two-pass approach to ensure accurate measurements:

1. **Pass 1**: Clean timing measurements with no monitoring overhead (50 iterations each)
2. **Pass 2**: Resource monitoring for CPU and memory usage (50 iterations each)

This separation prevents the monitoring overhead from affecting timing accuracy while still collecting comprehensive resource usage statistics.

### Test Environment
- **Template**: `mjml/testdata/austin-layout-from-mjml-io.mjml`
- **Iterations**: 50 per tool per measurement type
- **Platform**: macOS (darwin)
- **Timing**: High-resolution timing using Perl's `Time::HiRes` module
- **Resource Monitoring**: Process monitoring with 10ms sampling rate

## Latest Results

| Tool        | 50x Total (ms)      | Avg (ms)    | Max RAM (MB) | Avg CPU (%) |
|-------------|---------------------|-------------|-------------|-------------|
| gomjml      |                 162 |           3 |           2 |           0 |
| mrml        |                  96 |           1 |           1 |           0 |
| mjml (JS)   |               13530 |         270 |          90 |        20.9 |

### Key Insights

1. **Native Performance**: Both Go (gomjml) and Rust (mrml) implementations significantly outperform the JavaScript version
2. **Memory Efficiency**: Native implementations use minimal memory (1-2MB) vs JavaScript (90MB)
3. **CPU Usage**: Native implementations have negligible CPU overhead vs JavaScript (20.9% average)
4. **Production Readiness**: For high-throughput email generation, native implementations are essential

## Running Benchmarks

To run the benchmarks yourself:

```bash
# Run with ASCII table output
./bench-austin.sh

# Run with markdown table output
./bench-austin.sh --markdown

# Get help
./bench-austin.sh --help
```

## Benchmark Evolution

The benchmarking script has evolved to address timing accuracy issues:

- **Initial Implementation**: Simple timing with resource monitoring caused measurement interference
- **Problem**: Monitoring overhead affected timing results (2ms→22ms degradation)
- **Solution**: Two-pass methodology separating timing from resource monitoring
- **Enhancement**: Added markdown output support for documentation integration

This ensures both accurate timing measurements and comprehensive resource usage statistics.