# Run Go Engine Performance Test
locust -f performance-tests/go_engine_load.py --host http://localhost:8080 --headless -u 10 -r 2 -t 30s --html performance-report-go.html --csv performance-stats-go