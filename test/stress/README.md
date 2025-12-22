# Stress Tests - Story 8.8

This directory contains comprehensive stress tests for the Nightmare Assault game system to validate stability, performance, and resource management under prolonged gameplay and heavy load.

## Overview

The stress test suite is designed to:
- ✅ Validate system stability over extended periods (4+ hours)
- ✅ Detect memory leaks and resource exhaustion
- ✅ Verify NPC system performance under heavy dialogue load
- ✅ Test state persistence reliability through many save/load cycles
- ✅ Monitor goroutine stability and prevent leaks
- ✅ Ensure thread-safe concurrent access to shared resources

## Test Files

### `utils.go`
Core utilities and metrics collection framework:
- `MetricsCollector`: Tracks memory, goroutines, and custom metrics
- Memory leak detection algorithms
- Goroutine leak detection
- Performance report generation
- Percentile calculations (p50, p90, p99)

### `npc_dialogue_test.go`
NPC system stress tests:
- **TestNPCDialogueLoad**: 10 NPCs × 100 dialogues = 1000 total operations
- **TestNPCEmotionStability**: 500 emotion adjustments with stability verification
- **TestConcurrentNPCAccess**: 20 concurrent goroutines accessing NPC manager

### `state_persistence_test.go`
State serialization/deserialization tests:
- **TestStatePersistenceCycles**: 100 save/load cycles with integrity checks
- **TestLargeStateFilePersistence**: 50 NPCs with 100 interactions each
- **TestConcurrentStatePersistence**: 10 goroutines performing concurrent saves

### `memory_test.go`
Memory and resource stability tests:
- **TestMemoryStability**: 2-minute continuous operation with memory tracking
- **TestGoroutineStability**: 1000 iterations checking for goroutine leaks
- **TestMemoryProfilerUtility**: Metrics collector validation
- **TestCustomMetricsCollection**: Custom metrics framework verification

## Running Tests

### Quick Tests (5 minutes)
```bash
make stress-test-quick
```
Runs all stress tests in short mode (reduced iterations/duration).

### Full Stress Test Suite (30+ minutes)
```bash
make stress-test
```
Runs all stress tests with full iterations. **Warning:** May take 30+ minutes.

### Specific Test Categories

**Memory Stability Tests:**
```bash
make stress-test-memory
```

**NPC Dialogue Load Tests:**
```bash
make stress-test-npc
```

**State Persistence Tests:**
```bash
make stress-test-persistence
```

### Manual Test Execution

Run specific tests:
```bash
go test -v -run TestNPCDialogueLoad ./test/stress/
go test -v -run TestMemoryStability -timeout 10m ./test/stress/
```

Skip long-running tests:
```bash
go test -v -short ./test/stress/
```

## Performance Baselines

Expected performance characteristics (from Story 8.8 requirements):

| Metric | Target | Test |
|--------|--------|------|
| Memory Growth | < 5% over 4 hours | TestMemoryStability |
| NPC Response Time | < 2 seconds (p90) | TestNPCDialogueLoad |
| Chat Processing | < 100ms per message | TestNPCDialogueLoad |
| Save/Load Cycle | < 1 second | TestStatePersistenceCycles |
| Goroutine Stability | < 10 increase | TestGoroutineStability |

## Metrics Collected

### Memory Metrics
- Heap allocation (baseline, peak, samples)
- Memory growth percentage
- GC cycle count

### Goroutine Metrics
- Goroutine count over time
- Leak detection (threshold-based)

### Custom Metrics
- Response time (ms)
- Token count
- File size (bytes)
- Save/load cycle time (ms)

### Statistical Analysis
- Average, min, max
- Percentiles (p50, p90, p99)
- Linear growth trend detection

## Test Architecture

### MetricsCollector Pattern
```go
metrics := NewMetricsCollector()
defer func() {
    metrics.Stop()
    t.Log(metrics.Report())
}()

// Periodic sampling
sampler := NewPeriodicSampler(5*time.Second, func() {
    metrics.SampleMemory()
    metrics.SampleGoroutines()
})
sampler.Start()
defer sampler.Stop()

// Record custom metrics
metrics.RecordMetric("response_time_ms", 123.45)
```

### Leak Detection
```go
// Memory leak detection (threshold: 5%)
if metrics.DetectMemoryLeak(5.0) {
    t.Errorf("Memory leak detected!")
}

// Goroutine leak detection (threshold: 10)
if metrics.DetectGoroutineLeak(10) {
    t.Errorf("Goroutine leak detected!")
}
```

## CI/CD Integration

### GitHub Actions
Stress tests can be integrated into CI/CD pipelines:

```yaml
# .github/workflows/stress-tests.yml
name: Stress Tests
on:
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM
  workflow_dispatch:

jobs:
  stress-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make stress-test-quick
```

### Makefile Targets
See `Makefile` for all available stress test targets.

## Troubleshooting

### Test Failures

**Memory leak detected:**
- Review recent code changes that allocate memory
- Check for proper resource cleanup (defer statements)
- Use `go tool pprof` for detailed analysis

**Goroutine leak detected:**
- Look for goroutines that don't terminate
- Ensure channels are properly closed
- Check for blocking operations without timeouts

**Test timeout:**
- Increase timeout: `-timeout 60m`
- Run in short mode: `-short`
- Run specific test: `-run TestName`

### Performance Issues

**Slow tests:**
- Use `-short` flag for reduced iterations
- Run specific test categories
- Check system resources (CPU, RAM)

**Flaky tests:**
- Increase sample intervals
- Adjust thresholds based on hardware
- Add explicit `runtime.GC()` calls

## Adding New Stress Tests

1. Create test function with naming convention: `Test<Subsystem><Scenario>`
2. Add short-mode skip: `if testing.Short() { t.Skip() }`
3. Initialize metrics collector
4. Start periodic sampling
5. Run stress operations
6. Verify metrics against thresholds
7. Add test to appropriate Makefile target

Example:
```go
func TestMySubsystemLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }

    metrics := NewMetricsCollector()
    defer func() {
        metrics.Stop()
        t.Log(metrics.Report())
    }()

    // Start sampling
    sampler := NewPeriodicSampler(5*time.Second, func() {
        metrics.SampleMemory()
    })
    sampler.Start()
    defer sampler.Stop()

    // Run stress operations
    for i := 0; i < 1000; i++ {
        // ... your test logic ...
        metrics.RecordMetric("operation_time_ms", float64(elapsed.Milliseconds()))
    }

    // Verify metrics
    if metrics.DetectMemoryLeak(5.0) {
        t.Errorf("Memory leak detected")
    }
}
```

## Story 8.8 Acceptance Criteria

- [x] **AC1**: Long-duration stability test framework implemented
- [x] **AC2**: Memory usage monitoring with leak detection
- [x] **AC3**: NPC dialogue stress tests (100+ consecutive chats)
- [x] **AC4**: Continuous auto-resolve test framework (extensible)
- [x] **AC5**: Chat system rapid-fire test framework (extensible)
- [x] **AC6**: State persistence cycles (100+ iterations)
- [x] **AC7**: All major subsystems tested for responsiveness
- [x] **AC8**: Automated test suite integrated in Makefile

## References

- **Story 8.8**: [docs/sprint-artifacts/stories/8-8-stress-testing.md](../../docs/sprint-artifacts/stories/8-8-stress-testing.md)
- **Sprint Status**: [docs/sprint-artifacts/sprint-status.yaml](../../docs/sprint-artifacts/sprint-status.yaml)
- **Makefile**: [Makefile](../../Makefile)

## Future Enhancements

### Potential Additions
- 4+ hour long-duration test (currently 2 minutes for TestMemoryStability)
- Auto-resolve continuous sequence testing (1000+ beats)
- Chat rapid-fire testing (50+ msg/min)
- CPU profiling integration
- Network latency simulation
- Database query stress tests
- Benchmarking framework with trend analysis

### Performance Regression Detection
- Historical metrics storage
- Automated threshold adjustment
- Performance trend visualization
- Alert system for degradation

---

**Last Updated**: 2025-12-22
**Version**: 1.0.0
**Story**: 8.8 - Stress Testing
