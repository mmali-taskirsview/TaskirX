# TaskirX Next.js Dashboard

Modern, responsive admin dashboard for TaskirX v3.0 Ad Exchange Platform.

## 🎯 Features

- **Real-time Analytics**: Live performance metrics and charts
- **Campaign Management**: Create, edit, and monitor ad campaigns
- **AI Services Dashboard**: Monitor Fraud Detection, Ad Matching, and Bid Optimization
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile
- **Dark Mode Ready**: Built-in theme support
- **TypeScript**: Fully typed for better DX
- **Tailwind CSS**: Modern utility-first styling

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ installed
- Backend services running (NestJS, Go, Python AI agents)

### Installation

```powershell
# Install dependencies
cd C:\TaskirX\next-dashboard
npm install

# Start development server
npm run dev
```

Dashboard will be available at: **http://localhost:3001**

## 📦 Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript 5.3
- **Styling**: Tailwind CSS 3.4
- **Icons**: Lucide React
- **Charts**: Recharts
- **HTTP Client**: Axios
- **State Management**: Zustand

## 🏗️ Project Structure

```
next-dashboard/
├── app/
│   ├── dashboard/          # Main dashboard pages
│   │   ├── page.tsx        # Overview dashboard
│   │   ├── campaigns/      # Campaign management
│   │   ├── analytics/      # Analytics & reporting
│   │   ├── fraud/          # Fraud detection dashboard
│   │   ├── optimization/   # Bid optimization dashboard
│   │   └── settings/       # User settings
│   ├── layout.tsx          # Root layout
│   ├── page.tsx            # Landing page
│   └── globals.css         # Global styles
├── components/             # Reusable UI components
│   ├── button.tsx
│   ├── card.tsx
│   ├── header.tsx
│   └── sidebar.tsx
├── lib/
│   ├── api.ts             # API client
│   └── utils.ts           # Utility functions
├── next.config.js         # Next.js configuration
├── tailwind.config.ts     # Tailwind configuration
└── package.json
```

## 🔌 API Integration

The dashboard connects to multiple backend services:

- **NestJS Backend** (3000): Main API, auth, campaigns
- **Go Bidding Engine** (8080): Real-time bidding
- **Fraud Detection** (6001): ML-powered fraud analysis
- **Ad Matching** (6002): Intelligent ad matching
- **Bid Optimization** (6003): Thompson Sampling optimization

API proxying is configured in `next.config.js` to avoid CORS issues.

## 📊 Pages

### Dashboard Home (`/dashboard`)
- Real-time performance metrics
- Revenue, impressions, clicks, CTR
- AI services status
- Recent activity feed

### Campaigns (`/dashboard/campaigns`)
- List all campaigns
- Create/edit campaigns
- Campaign performance
- Budget tracking

### Analytics (`/dashboard/analytics`)
- Performance charts
- Conversion tracking
- Revenue analytics
- Detailed reporting

### Fraud Detection (`/dashboard/fraud`)
- Fraud metrics
- Real-time threat monitoring
- ML model performance
- IP blacklists

### Optimization (`/dashboard/optimization`)
- Bid optimization insights
- Thompson Sampling status
- Budget pacing
- ROI improvements

## 🎨 Customization

### Colors

Edit `tailwind.config.ts` to customize the color scheme:

```typescript
theme: {
  extend: {
    colors: {
      primary: 'hsl(var(--primary))',
      // Add your colors
    }
  }
}
```

### Branding

Update logo and branding in:
- `components/sidebar.tsx`
- `components/header.tsx`
- `app/page.tsx`

## 🔧 Environment Variables

Create `.env.local`:

```env
# API Endpoints
BACKEND_URL=http://localhost:4000
NEXT_PUBLIC_BACKEND_URL=http://localhost:4000/api
NEXT_PUBLIC_FRAUD_URL=http://localhost:6001/api
NEXT_PUBLIC_MATCHING_URL=http://localhost:6002/api
NEXT_PUBLIC_OPTIMIZATION_URL=http://localhost:6003/api

# Auth
NEXTAUTH_SECRET=your-secret-key-here

# Features
NEXT_PUBLIC_ENABLE_ANALYTICS=true
NEXT_PUBLIC_ENABLE_REALTIME=true
```

**Note**: If you run the stack with `docker-compose`, use `http://localhost:3000/api` for `NEXT_PUBLIC_BACKEND_URL`.

## 🧪 Testing

```powershell
# Type checking
npm run type-check

# Linting
npm run lint

# Build production
npm run build

# Start production server
npm start
```

## 🚀 Production Deployment

### Build for Production

```powershell
npm run build
```

### Deploy to Vercel

```powershell
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel
```

### Deploy to Docker

```powershell
# Build Docker image
docker build -t taskir-dashboard .

# Run container
docker run -p 3001:3001 taskir-dashboard
```

## 📈 Performance

- **Lighthouse Score**: 95+
- **First Contentful Paint**: <1s
- **Time to Interactive**: <2s
- **Bundle Size**: ~200KB gzipped

## 🔐 Security

- JWT authentication
- CSRF protection
- XSS prevention
- Secure HTTP-only cookies
- Rate limiting

## 📝 Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm start` - Start production server
- `npm run lint` - Lint code
- `npm run type-check` - TypeScript type checking

## 🐛 Troubleshooting

### Port 3001 already in use

```powershell
# Kill process on port 3001
netstat -ano | findstr :3001
taskkill /PID <PID> /F
```

### Dependencies not installing

```powershell
# Clear npm cache
npm cache clean --force
rm -rf node_modules package-lock.json
npm install
```

### Build errors

```powershell
# Clear Next.js cache
rm -rf .next
npm run build
```

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Test thoroughly
4. Submit a pull request

## 📄 License

Proprietary - TaskirX v3.0

---

**Status**: ✅ Core Dashboard Complete  
**Version**: 3.0.0  
**Last Updated**: January 28, 2026  

For questions, contact: admin@taskir.com
