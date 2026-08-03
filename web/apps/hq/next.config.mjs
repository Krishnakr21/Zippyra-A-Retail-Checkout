/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ["@zippyra/ui", "@zippyra/types", "@zippyra/api-client", "@zippyra/hooks"],
  async rewrites() {
    return [
      { source: '/v1/chain-hq/:path*', destination: 'http://localhost:8096/v1/chain-hq/:path*' },
      { source: '/v1/catalog/:path*', destination: 'http://localhost:8083/v1/catalog/:path*' },
      { source: '/v1/store/:path*', destination: 'http://localhost:8082/v1/store/:path*' },
      { source: '/v1/analytics/:path*', destination: 'http://localhost:8092/v1/analytics/:path*' },
    ];
  },
};

export default nextConfig;
