# 🚀 JavaScript SDK - Build & Test Guide

## Quick Start

### 1. Install Dependencies

```bash
cd sdks/javascript
npm install
```

This will install:
- TypeScript 5.3.3
- Webpack 5.89.0
- ts-loader 9.5.1
- Jest 29.7.0 (for testing)

### 2. Build the SDK

```bash
# Production build (minified)
npm run build

# Development build (watch mode)
npm run dev

# Type checking only
npm run type-check
```

**Output**: `dist/adx-sdk.js` (UMD bundle, ~35KB minified + gzipped)

### 3. Test the SDK

#### Option A: Use the Example HTML File

```bash
# Make sure your TaskirX server is running
cd ../../backend
npm start

# In another terminal, open the example
cd ../sdks/javascript
# Open example.html in your browser
start example.html  # Windows
# or
open example.html   # macOS
# or
xdg-open example.html  # Linux
```

The example page includes:
- SDK initialization
- Banner ad display
- Native ad display
- Video ad display
- Real-time status updates

#### Option B: Manual Testing

Create a simple HTML file:

```html
<!DOCTYPE html>
<html>
<head>
    <title>AdxSDK Test</title>
</head>
<body>
    <h1>AdxSDK Test</h1>
    
    <div id="banner-ad" style="width: 320px; height: 50px; border: 1px solid #ccc;"></div>
    
    <script src="dist/adx-sdk.js"></script>
    <script>
        // Initialize
        AdxSDK.init({
            publisherId: 'publisher@test.com',
            apiEndpoint: 'http://localhost:3000',
            enableDebug: true
        });
        
        // Show banner ad
        AdxSDK.showBanner({
            placementId: 'banner-home',
            containerId: 'banner-ad',
            width: 320,
            height: 50
        }).then(() => {
            console.log('Ad loaded!');
        }).catch(error => {
            console.error('Ad failed:', error);
        });
    </script>
</body>
</html>
```

#### Option C: Run Unit Tests

```bash
# Run Jest tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

## 🔍 Troubleshooting

### Build Issues

**Problem**: `Cannot find module 'typescript'`
```bash
# Solution: Reinstall dependencies
rm -rf node_modules package-lock.json
npm install
```

**Problem**: Webpack build fails
```bash
# Solution: Check Node.js version (need 18+)
node --version

# Update Node.js if needed
nvm install 20
nvm use 20
```

### Runtime Issues

**Problem**: "AdxSDK is not defined"
- **Solution**: Make sure `dist/adx-sdk.js` is loaded before using the SDK
- Check browser console for loading errors
- Verify the script path is correct

**Problem**: "No ad available" (no fill)
- **Solution**: Make sure campaigns exist in the database
- Check backend logs for bid request/response
- Verify placement ID matches campaign targeting

**Problem**: CORS errors
- **Solution**: Backend server must allow CORS for your domain
- Check backend `cors` middleware configuration
- Use same domain or configure CORS properly

**Problem**: Network errors
- **Solution**: Ensure backend server is running on `http://localhost:3000`
- Check backend health: `curl http://localhost:3000/health`
- Verify API endpoint in SDK init config

## 📊 Testing Checklist

Before deployment, verify:

- [ ] **SDK builds without errors**: `npm run build`
- [ ] **Type checking passes**: `npm run type-check`
- [ ] **Bundle size is acceptable**: Check `dist/adx-sdk.js` size
- [ ] **Banner ads load and display**: Test with example.html
- [ ] **Native ads render correctly**: Check custom templates
- [ ] **Video ads play**: Test autoplay, controls, muted options
- [ ] **Impression tracking works**: Check backend logs
- [ ] **Click tracking works**: Click ads and verify backend tracking
- [ ] **Viewability tracking works**: Scroll ads in/out of view
- [ ] **User ID persists**: Check localStorage
- [ ] **Device detection works**: Test on mobile, tablet, desktop
- [ ] **Error handling works**: Test with invalid placement IDs
- [ ] **Debug logging works**: Check browser console

## 🎯 Integration Testing

### Test with Real Backend

1. **Start backend server**:
```bash
cd ../../backend
npm start
```

2. **Login and get token** (optional):
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "publisher@test.com",
    "password": "password123"
  }'
```

3. **Create test campaign** (if needed):
```bash
# Use Postman or similar tool
POST http://localhost:3000/api/campaigns
Authorization: Bearer <your-token>
Content-Type: application/json

{
  "name": "Test Campaign",
  "budget": 1000,
  "dailyBudget": 100,
  "bidAmount": 2.50,
  "targeting": {
    "placements": ["banner-home", "native-feed", "video-pre-roll"]
  },
  "creative": {
    "format": "banner",
    "imageUrl": "https://via.placeholder.com/320x50",
    "clickUrl": "https://example.com"
  }
}
```

4. **Test SDK with example.html**:
- Open `example.html` in browser
- Click "Initialize SDK"
- Click "Load Banner Ad"
- Verify ad displays
- Check browser console for logs
- Check backend logs for bid requests

### Test with Mock Server

If you don't have the backend running, you can mock responses:

```javascript
// Mock fetch for testing
const originalFetch = window.fetch;
window.fetch = function(url, options) {
    if (url.includes('/api/rtb/bid-request')) {
        return Promise.resolve({
            ok: true,
            json: () => Promise.resolve({
                id: 'mock-bid-123',
                seatbid: [{
                    bid: [{
                        id: 'bid-1',
                        impid: 'imp-1',
                        price: 2.50,
                        adm: 'https://via.placeholder.com/320x50'
                    }]
                }]
            })
        });
    }
    return originalFetch(url, options);
};
```

## 📦 Production Build

### Optimize for Production

```bash
# Build with optimizations
npm run build

# Check bundle size
ls -lh dist/adx-sdk.js

# Analyze bundle (optional)
npm run analyze  # If you add webpack-bundle-analyzer
```

### CDN Distribution

1. **Upload to CDN**:
```bash
# Example: AWS S3
aws s3 cp dist/adx-sdk.js s3://your-bucket/adx-sdk/v1.0.0/adx-sdk.js
aws s3 cp dist/adx-sdk.js s3://your-bucket/adx-sdk/latest/adx-sdk.js
```

2. **Set cache headers**:
```bash
aws s3api put-object-acl \
  --bucket your-bucket \
  --key adx-sdk/v1.0.0/adx-sdk.js \
  --acl public-read \
  --cache-control "max-age=31536000"
```

3. **Test CDN URL**:
```html
<script src="https://cdn.yourdomain.com/adx-sdk/latest/adx-sdk.js"></script>
```

### NPM Distribution

1. **Update package.json**:
```json
{
  "name": "@taskirx/sdk",
  "version": "1.0.0",
  "main": "dist/adx-sdk.js",
  "types": "dist/index.d.ts",
  "files": ["dist"]
}
```

2. **Publish to npm**:
```bash
npm login
npm publish --access public
```

3. **Install from npm**:
```bash
npm install @taskirx/sdk
```

## 🔧 Development Workflow

### Watch Mode

```bash
# Auto-rebuild on file changes
npm run dev
```

Keep this running while developing. It will:
- Watch `src/**/*.ts` files
- Rebuild on changes
- Output to `dist/adx-sdk.js`

### Code Style

```bash
# Run linter
npm run lint

# Fix linting issues
npm run lint:fix

# Format code
npm run format  # If you add prettier
```

### Debugging

Enable debug mode:
```javascript
AdxSDK.init({
    publisherId: 'test',
    enableDebug: true  // This logs to console
});
```

Check browser console for:
- Bid requests/responses
- Impression tracking
- Click tracking
- Viewability events
- Errors and warnings

### Browser DevTools

Use Chrome DevTools:
1. Open DevTools (F12)
2. Go to Network tab
3. Filter by "Fetch/XHR"
4. Watch for requests to `/api/rtb/`
5. Inspect request/response payloads

## 📈 Performance Testing

### Load Testing

Test SDK performance with multiple ads:

```html
<script>
    // Load 10 banner ads
    for (let i = 0; i < 10; i++) {
        AdxSDK.showBanner({
            placementId: 'banner-home',
            containerId: `ad-${i}`,
            width: 320,
            height: 50
        });
    }
</script>
```

### Memory Testing

Monitor memory usage:
1. Open Chrome DevTools
2. Go to Performance tab
3. Record page load
4. Load multiple ads
5. Check memory snapshots

### Network Testing

Test on slow networks:
1. Open Chrome DevTools
2. Go to Network tab
3. Select "Slow 3G" throttling
4. Test ad loading

## ✅ Ready for Production

When all tests pass:
- [x] Build succeeds
- [x] No TypeScript errors
- [x] Bundle size < 50KB
- [x] All ad formats work
- [x] Tracking works
- [x] Error handling works
- [x] Browser compatibility tested
- [x] Mobile tested
- [x] Performance acceptable

You're ready to deploy! 🚀
