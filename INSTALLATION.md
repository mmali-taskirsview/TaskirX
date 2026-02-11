# Installation Guide for TaskirX

## System Requirements

### Required Software
- **Node.js**: Version 18.0.0 or higher
- **npm**: Version 9.0.0 or higher (comes with Node.js)
- **MongoDB**: Version 7.0 or higher
- **Redis**: Version 7.0 or higher

### Operating System
- Windows 10/11
- Linux (Ubuntu 20.04+)
- macOS 12+

## Step-by-Step Installation

### 1. Install Node.js

#### Windows
1. Download Node.js from https://nodejs.org/
2. Download the LTS version (currently 18.x or 20.x)
3. Run the installer (.msi file)
4. Follow installation wizard (accept defaults)
5. Verify installation:
   ```powershell
   node --version
   npm --version
   ```

#### Linux (Ubuntu/Debian)
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
node --version
npm --version
```

#### macOS
```bash
# Using Homebrew
brew install node@18
node --version
npm --version
```

### 2. Install MongoDB

#### Windows
1. Download from https://www.mongodb.com/try/download/community
2. Choose Windows MSI installer
3. Run installer and select "Complete" installation
4. Install MongoDB as a Windows Service (recommended)
5. Verify installation:
   ```powershell
   mongosh --version
   ```

#### Linux (Ubuntu)
```bash
# Import MongoDB GPG key
wget -qO - https://www.mongodb.org/static/pgp/server-7.0.asc | sudo apt-key add -

# Add MongoDB repository
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu $(lsb_release -cs)/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list

# Install MongoDB
sudo apt-get update
sudo apt-get install -y mongodb-org

# Start MongoDB
sudo systemctl start mongod
sudo systemctl enable mongod
```

#### macOS
```bash
# Using Homebrew
brew tap mongodb/brew
brew install mongodb-community@7.0

# Start MongoDB
brew services start mongodb-community@7.0
```

### 3. Install Redis

#### Windows
Option 1 - Using WSL (Recommended):
```powershell
# Install WSL first
wsl --install

# In WSL terminal
sudo apt-get update
sudo apt-get install redis-server
sudo service redis-server start
```

Option 2 - Memurai (Windows native Redis alternative):
1. Download from https://www.memurai.com/get-memurai
2. Install and run as Windows service

#### Linux (Ubuntu)
```bash
sudo apt-get update
sudo apt-get install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
redis-cli ping  # Should return "PONG"
```

#### macOS
```bash
brew install redis
brew services start redis
redis-cli ping  # Should return "PONG"
```

### 4. Install Project Dependencies

```bash
# Navigate to project directory
cd c:\taskirx

# Install all dependencies
npm install

# This will install:
# - Express.js (web framework)
# - Mongoose (MongoDB ODM)
# - ioredis (Redis client)
# - JWT (authentication)
# - Winston (logging)
# - And 15+ other packages
```

### 5. Configure Environment

Create `.env` file in project root:

```bash
# Copy the example file
copy .env.example .env

# Edit .env with your settings
notepad .env
```

**Minimum Required Configuration:**
```env
NODE_ENV=development
PORT=3000
BASE_URL=http://localhost:3000

MONGODB_URI=mongodb://localhost:27017/taskirx

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=change-this-to-a-random-secret-key-min-32-chars
```

### 6. Verify Installation

#### Check Services
```powershell
# MongoDB
mongosh --eval "db.version()"

# Redis
redis-cli ping

# Node.js
node --version
npm --version
```

#### Test Database Connections
```powershell
# Start the server
npm run dev

# Check health endpoint (in another terminal)
curl http://localhost:3000/health
```

Expected output:
```json
{
  "status": "ok",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

## Common Installation Issues

### Issue: npm command not found
**Solution:**
- Restart PowerShell/terminal after Node.js installation
- Check PATH: `echo $env:PATH` (should include Node.js path)
- Reinstall Node.js with "Add to PATH" option selected

### Issue: MongoDB connection refused
**Solution:**
```powershell
# Windows - Start MongoDB service
net start MongoDB

# Check status
sc query MongoDB

# Or start manually
"C:\Program Files\MongoDB\Server\7.0\bin\mongod.exe" --dbpath="C:\data\db"
```

### Issue: Redis connection refused
**Solution:**
```powershell
# If using WSL
wsl
sudo service redis-server start
sudo service redis-server status

# Test connection
redis-cli ping
```

### Issue: Port 3000 already in use
**Solution:**
```powershell
# Find process using port 3000
netstat -ano | findstr :3000

# Kill the process (replace <PID> with actual PID)
taskkill /PID <PID> /F

# Or change PORT in .env file
PORT=3001
```

### Issue: Permission errors during npm install
**Solution:**
```bash
# Windows - Run PowerShell as Administrator
# Or fix npm permissions:
npm config set prefix "$env:APPDATA\npm"
```

### Issue: EACCES errors on Linux/Mac
**Solution:**
```bash
sudo chown -R $USER:$USER ~/.npm
sudo chown -R $USER:$USER node_modules
```

## Development Tools (Optional)

### MongoDB GUI Clients
- **MongoDB Compass** (Official): https://www.mongodb.com/products/compass
- **Studio 3T**: https://studio3t.com/

### Redis GUI Clients
- **RedisInsight** (Official): https://redis.com/redis-enterprise/redis-insight/
- **Redis Commander**: `npm install -g redis-commander`

### API Testing Tools
- **Postman**: https://www.postman.com/downloads/
- **Insomnia**: https://insomnia.rest/download
- **VS Code REST Client**: Extension for VS Code

### Monitoring Tools
- **PM2** (Process Manager): `npm install -g pm2`
- **nodemon** (Auto-reload): Already included in dev dependencies

## Post-Installation Steps

1. **Create Admin User**
   ```bash
   # Register admin user via API
   curl -X POST http://localhost:3000/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@example.com",
       "password": "Admin1234!",
       "name": "Admin User",
       "company": "TaskirX"
     }'
   
   # Then manually update role in MongoDB
   mongosh
   use taskirx
   db.users.updateOne(
     {email: "admin@example.com"},
     {$set: {role: "admin"}}
   )
   ```

2. **Setup Indexes** (Automatic on first run)
   - Indexes are created automatically by Mongoose schemas
   - Check indexes: `db.campaigns.getIndexes()` in mongosh

3. **Configure Monitoring** (Production)
   - Setup Prometheus for metrics scraping
   - Configure alerts for health checks
   - Setup log aggregation (ELK stack, CloudWatch, etc.)

4. **Security Hardening** (Production)
   - Change JWT_SECRET to strong random value
   - Enable MongoDB authentication
   - Configure Redis password
   - Setup SSL/TLS certificates
   - Enable rate limiting
   - Configure firewall rules

## Next Steps

✅ Installation complete! Now you can:

1. **Start Development**: `npm run dev`
2. **Read Quick Start**: See `QUICKSTART.md`
3. **Study Architecture**: Review `ARCHITECTURE.md`
4. **Test APIs**: Import Postman collection
5. **Run Benchmarks**: `npm run benchmark`

## Getting Help

- Check logs in `logs/` directory
- Review error messages carefully
- Search GitHub issues
- Read documentation in `docs/` folder
- Check MongoDB/Redis logs

## System Architecture Verification

After installation, verify your setup matches requirements:

```bash
# Check Node.js version (should be 18+)
node --version

# Check dependencies
npm list --depth=0

# Check MongoDB version (should be 7.0+)
mongosh --eval "db.version()"

# Check Redis version (should be 7.0+)
redis-cli INFO server | grep redis_version

# Test server startup
npm run dev

# Run health check
curl http://localhost:3000/health
```

All checks should pass! 🚀
