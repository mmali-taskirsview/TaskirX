# Quick Start Guide - TaskirX

## The Issue You're Experiencing

**Problem**: When running `npm run dev` in PowerShell, any subsequent command triggers "Terminate batch job (Y/N)?" This is a PowerShell limitation with background processes.

**Solution**: Use **two separate windows** - one for the server, one for testing.

---

## How to Start & Test TaskirX

### Method 1: Using CMD Files (Simplest - RECOMMENDED)

**Step 1** - Double-click `start-server.cmd` in File Explorer
  - This opens a new window with the server running
  - Keep this window open

**Step 2** - Double-click `quick-test.cmd` in File Explorer  
  - This tests if the server is working
  - Shows health check results

**Step 3** - For full testing, open PowerShell and run:
```powershell
cd C:\taskirx
.\test-endpoints.ps1
```

---

### Method 2: Using Two PowerShell Windows

**PowerShell Window #1** (Server):
```powershell
cd C:\taskirx
npm run dev
```
**Keep this window open and running!**

**PowerShell Window #2** (Testing):
```powershell
cd C:\taskirx

# Quick test
Invoke-RestMethod http://localhost:3000/health

# Full test
.\test-endpoints.ps1
```

---

## Why "Terminate batch job" Appears

When `npm run dev` starts `nodemon`, it creates a persistent process. PowerShell intercepts EVERY new command in that terminal and asks if you want to kill the process first. This is normal PowerShell behavior.

**The fix**: Don't run commands in the same window as `npm run dev`. Use a separate window.

---

## Quick Commands Reference

### Start Server (Choose ONE):
```bash
# Option A: CMD file (easiest)
start-server.cmd

# Option B: PowerShell
npm run dev
```

### Test Server (In a NEW window):
```powershell
# Simple health check
Invoke-RestMethod http://localhost:3000/health

# Login and get token
$body = '{"email":"advertiser@example.com","password":"password123"}'
$login = Invoke-RestMethod http://localhost:3000/api/auth/login -Method Post -ContentType "application/json" -Body $body
$token = $login.token

# Get campaigns
$headers = @{"Authorization" = "Bearer $token"}
Invoke-RestMethod http://localhost:3000/api/campaigns -Headers $headers

# Get analytics
Invoke-RestMethod http://localhost:3000/api/analytics/dashboard -Headers $headers
```

---

## Stop the Server

**If server is running in a terminal:**
- Press `Ctrl+C` (you'll see "Terminate batch job" - press `Y`)

**If you need to force-stop all Node processes:**
```powershell
Stop-Process -Name node -Force
```

---

## Your Platform Status

✅ **Server code is working correctly!** The logs show:
- ✅ MongoDB connected successfully
- ✅ Database indexes created  
- ✅ Server started on port 3000
- ⚠️ Redis disabled (optional - platform works without it)

The only "issue" is PowerShell's normal behavior when running background processes.

---

## Next Steps

1. **Start**: Run `start-server.cmd` (or `npm run dev` in one window)
2. **Test**: Run `quick-test.cmd` (or `.\test-endpoints.ps1` in another window)
3. **Develop**: Import `postman-collection.json` into Postman
4. **Deploy**: See `docs/DEPLOYMENT.md`
