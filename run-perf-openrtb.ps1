# Run OpenRTB Performance Test (Locust)
locust -f performance-tests/openrtb_load.py --host http://localhost:8082 --headless -u 10 -r 2 -t 30s --html performance-report-openrtb.html --csv performance-stats-openrtb
