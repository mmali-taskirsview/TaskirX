# Performance Tests

Load tests for TaskirX endpoints using Locust.

## Scenarios
- **SSP Auction**: `POST /api/ssp/auction` (OpenRTB 2.5 style payload)
- **DSP Bid**: `POST /api/dsp/bid`
- **Analytics Dashboard**: `GET /api/analytics/dashboard` (optional auth)

## Prerequisites
- Python 3.11+
- Install dependencies:
  ```powershell
  pip install -r requirements.txt
  ```
- **Recommended**: Seed the database with performance entities (Publishers/Ad Units) to avoid 404s.
  ```powershell
  # Must be run against your Postgres DB
  psql -h localhost -U postgres -d taskirx -f ../scripts/seed-perf-data.sql
  ```

## Running
Use the helper script (auto-headless unless `-Headed`):
```powershell
# from repo root or performance-tests/
powershell -File performance-tests/run-perf-tests.ps1 -TargetHost http://localhost:3000 -Users 10 -SpawnRate 2 -RunTime "30s"
```
- Default host: `http://localhost:3000` (override with `-TargetHost` or env `LOCUST_HOST`).
- Optional auth for analytics: set env `LOCUST_TOKEN` to a JWT; the script will send `Authorization: Bearer <token>`.

Direct Locust (UI):
```powershell
locust -f performance-tests/locustfile.py --host http://localhost:3000
```

## Notes
- If the target services aren’t running, Locust will report status `0` (connection errors).
- Analytics may return 401 without `LOCUST_TOKEN`; that’s expected unless you supply a token.
