/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  images: {
    domains: ['localhost'],
  },
  async rewrites() {
    return [
      {
        source: '/api/backend/:path*',
        destination: process.env.NEXT_PUBLIC_API_URL ? `${process.env.NEXT_PUBLIC_API_URL}/api/:path*` : 'http://localhost:3000/api/:path*',
      },
      {
        source: '/api/bidding/:path*',
        destination: process.env.NEXT_PUBLIC_BIDDING_URL ? `${process.env.NEXT_PUBLIC_BIDDING_URL}/:path*` : 'http://localhost:8080/api/:path*',
      },
      {
        source: '/api/fraud/:path*',
        destination: 'http://localhost:6001/api/:path*',
      },
      {
        source: '/api/matching/:path*',
        destination: 'http://localhost:6002/api/:path*',
      },
      {
        source: '/api/optimization/:path*',
        destination: 'http://localhost:6003/api/:path*',
      },
    ];
  },
}

module.exports = nextConfig
