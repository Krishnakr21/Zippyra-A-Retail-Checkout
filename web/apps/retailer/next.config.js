/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@zippyra/ui', '@zippyra/api-client', '@zippyra/auth', '@zippyra/hooks', '@zippyra/types'],
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.amazonaws.com',
      },
      {
        protocol: 'https',
        hostname: '**.cloudfront.net',
      },
    ],
  },
  async rewrites() {
    return [
      { source: '/v1/cart/:path*', destination: 'http://localhost:8084/v1/cart/:path*' },
      { source: '/v1/store/:path*', destination: 'http://localhost:8082/v1/store/:path*' },
      { source: '/v1/catalog/:path*', destination: 'http://localhost:8083/v1/catalog/:path*' },
      { source: '/v1/inventory/:path*', destination: 'http://localhost:8090/v1/inventory/:path*' },
      { source: '/v1/analytics/:path*', destination: 'http://localhost:8092/v1/analytics/:path*' },
      { source: '/v1/transfer/:path*', destination: 'http://localhost:8100/v1/transfer/:path*' },
      { source: '/v1/transfers/:path*', destination: 'http://localhost:8100/v1/transfers/:path*' },
      { source: '/v1/device/:path*', destination: 'http://localhost:8102/v1/device/:path*' },
      { source: '/v1/devices/:path*', destination: 'http://localhost:8102/v1/devices/:path*' },
      { source: '/v1/support/:path*', destination: 'http://localhost:8093/v1/support/:path*' },
      { source: '/v1/auth/:path*', destination: 'http://localhost:8080/v1/auth/:path*' },
      { source: '/v1/retailer-auth/:path*', destination: 'http://localhost:8094/v1/retailer-auth/:path*' },
      { source: '/v1/order/:path*', destination: 'http://localhost:8086/v1/order/:path*' },
      { source: '/v1/payment/:path*', destination: 'http://localhost:8085/v1/payment/:path*' },
      { source: '/v1/compliance/:path*', destination: 'http://localhost:8098/v1/compliance/:path*' },
      { source: '/v1/dpdp/:path*', destination: 'http://localhost:8098/v1/dpdp/:path*' },
      { source: '/v1/qc/:path*', destination: 'http://localhost:8099/v1/qc/:path*' },
      { source: '/v1/warehouse/:path*', destination: 'http://localhost:8091/v1/warehouse/:path*' },
      { source: '/v1/grn/:path*', destination: 'http://localhost:8091/v1/grn/:path*' },
    ];
  },
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'Content-Security-Policy',
            value: `default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' https: data:; connect-src 'self' http://localhost:* ws://localhost:*;`,
          },
        ],
      },
    ];
  },
};

module.exports = nextConfig;
