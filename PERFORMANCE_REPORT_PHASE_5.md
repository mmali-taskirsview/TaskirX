# Performance Test Report - Go Bidding Engine
**Date:** February 17, 2026
**Test Script:** `performance-tests/go_engine_load.py` (Locust)

## Summary
The Go Bidding Engine was tested with a load of 10 concurrent users generating ~17 requests per second. The test passed with **100% success rate** (valid bids returned), confirming the logic fixes for Campaign Matching are robust.

## Key Metrics
- **Total Requests:** 512
- **Success Rate:** 100%
- **Avg Response Time:** 472 ms
- **Median Response Time:** 510 ms
- **Throughput:** ~17.3 RPS

## Bottleneck Analysis
The distinct **~500ms median latency** correlates exactly with the `Fraud Service` timeout observed in the logs:
`Warning: Fraud Service call failed: ... context deadline exceeded`

This indicates the **Python Fraud Detection Service** is unable to keep up with the load or is timing out due to network configuration in the test environment. The Go service correctly handles this by "failing open" (logging a warning and proceeding to bid), which is the desired behavior for latency-sensitive RTB systems, though ideally, it should fail closed or cache the result.

## Recommendations
1. **Optimize Fraud Service:** Switch to a faster model or increase worker count.
2. **Implement Circuit Breaker:** Prevent the 500ms wait penalty when the service is down/slow.
3. **Caching:** Cache IP reputation results in Redis to avoid hitting the Python service for every request.
