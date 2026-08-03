/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ['@zippyra/ui', '@zippyra/api-client', '@zippyra/auth', '@zippyra/hooks', '@zippyra/types'],
  async rewrites() {
    return [
      { source: '/v1/admin-auth/:path*', destination: 'http://localhost:8095/v1/admin-auth/:path*' },
      { source: '/v1/admin-store/:path*', destination: 'http://localhost:8097/v1/admin-store/:path*' },
      { source: '/v1/store/:path*', destination: 'http://localhost:8082/v1/store/:path*' },
      { source: '/v1/catalog/:path*', destination: 'http://localhost:8083/v1/catalog/:path*' },
      { source: '/v1/compliance/:path*', destination: 'http://localhost:8098/v1/compliance/:path*' },
      { source: '/v1/analytics/:path*', destination: 'http://localhost:8092/v1/analytics/:path*' },
    ];
  },
};

module.exports = nextConfig;
