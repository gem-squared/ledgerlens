/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Backend Go binary runs on PORT 8082 by default; rewrite /api/* to it in dev.
  async rewrites() {
    const backend = process.env.LEDGERLENS_BACKEND_URL || 'http://127.0.0.1:8082';
    return [
      {
        source: '/api/:path*',
        destination: `${backend}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
