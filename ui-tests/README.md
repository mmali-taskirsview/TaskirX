# TaskirX UI Tests

Playwright smoke tests for the TaskirX Next.js dashboard. The tests create users via the backend API and verify role-based redirects plus core page availability.

## Prerequisites

- Dashboard running at `http://localhost:3001`
- Backend API running at `http://localhost:3000/api`

Set env vars if needed:

- `DASHBOARD_BASE_URL`
- `API_BASE_URL`

## Run

```powershell
npm install
npx playwright install
npm test
```
