# Changelog

All notable changes to the TaskirX project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.2] - 2026-01-28

### 🔒 Security
- **CRITICAL:** Removed JWT secret fallback values in `auth.js` and `env.js`
- **HIGH:** Added environment variable validation on startup
- **HIGH:** Enabled Content Security Policy (CSP) with proper directives
- Enhanced .gitignore to prevent secret leakage
- Created comprehensive security checklist (400+ lines)

### ✨ Added
- Environment variable validator with detailed error messages
- Secure JWT secret generator script (`npm run generate-secret`)
- Unit tests for bidding engine (160+ lines)
- Unit tests for fraud detection (100+ lines)
- Unit tests for authentication middleware (90+ lines)
- Jest configuration and test setup
- Product roadmap documentation (ROADMAP.md)
- Backend structure documentation (BACKEND_STRUCTURE.md)
- Security checklist documentation (SECURITY_CHECKLIST.md)
- Comprehensive fixes documentation (FIXES_APPLIED.md)
- Security summary documentation (SECURITY_FIXES_SUMMARY.md)

### 🔧 Changed
- Updated Helmet security headers configuration
- Improved startup error messages
- Version bumped from 2.0.0 to 2.0.2
- Enhanced environment configuration validation

### 📚 Documentation
- Created 5 new comprehensive documentation files
- Documented all Phase 2 features in ROADMAP.md
- Clarified backend structure and entry points
- Added detailed security guidelines

### 🧪 Testing
- Initial test coverage: 15%
- 3 test suites with 350+ lines of tests
- Configured Jest with proper test environment
- Added test setup with environment mocking

---

## [2.0.0] - 2025-11-14

### ✨ Initial Release
- Complete RTB platform with OpenRTB 2.5 support
- Campaign management system
- Mobile attribution (6 MMP providers)
- JWT authentication & RBAC
- GDPR/CCPA compliance
- 3 Mobile SDKs (JavaScript, Android, iOS)
- Production logging & monitoring
- Docker support
- Comprehensive documentation (20+ files)
- 40+ API endpoints
- Swagger API documentation

### 🎯 Features
- Real-time bidding engine
- Second-price and first-price auctions
- Budget management & pacing
- Performance analytics
- Fraud detection
- Rate limiting
- Redis caching (optional)
- Sentry error tracking
- Health monitoring

---

## Security Advisories

### [2.0.1] - Skipped
Version 2.0.1 was skipped to go directly to 2.0.2 with all security fixes.

### [2.0.0] - Known Issues (RESOLVED in 2.0.2)
- JWT_SECRET had fallback to insecure default - **FIXED**
- CSP was disabled - **FIXED**
- No environment validation - **FIXED**
- No unit tests - **PARTIALLY FIXED** (15% coverage)

---

## Upgrade Guide

### From 2.0.0 to 2.0.2

1. **Pull latest changes**
   ```bash
   git pull origin main
   ```

2. **Install new dependencies** (no new dependencies, but ensure up to date)
   ```bash
   npm install
   ```

3. **Generate secure JWT secret**
   ```bash
   npm run generate-secret
   ```

4. **Update .env file**
   - Ensure JWT_SECRET is not a default value
   - Run validator to check: `npm start` (will error if invalid)

5. **Run tests**
   ```bash
   npm test
   ```

6. **Verify production readiness**
   ```bash
   npm run verify
   ```

7. **Review new documentation**
   - SECURITY_CHECKLIST.md
   - ROADMAP.md
   - FIXES_APPLIED.md

---

## Breaking Changes

### Version 2.0.2
- **Breaking:** Application now fails to start if JWT_SECRET is not set or is a default value
- **Breaking:** Content Security Policy is now enabled (may affect ad rendering)

**Migration:**
1. Set JWT_SECRET to a secure value (run `npm run generate-secret`)
2. Adjust CSP directives in `server.js` if needed for your ad formats

---

## Deprecation Notices

### Version 2.0.2
- `backend/server.js` is deprecated in favor of `backend/src/server.js`
- Default/fallback environment values are deprecated and will cause startup errors

### Planned for 3.0.0
- API versioning will be introduced (/api/v1/, /api/v2/)
- Microservices architecture for scale
- Breaking changes to SDK initialization

---

## Links
- [Repository](https://github.com/yourusername/taskirx)
- [Documentation](./README.md)
- [Security Policy](./SECURITY_CHECKLIST.md)
- [Roadmap](./ROADMAP.md)
