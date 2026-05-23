/** @type {import('next').NextConfig} */
const isStaticExport = process.env.NEXT_OUTPUT_MODE === 'export';

const nextConfig = {
  reactStrictMode: true,

  // STATIC EXPORT for production deploy (single-binary, embedded into Go).
  // In dev, we run a normal Next.js server with /api/* rewritten to the Go
  // backend so dev-on-3001 + go-on-8082 hot-reload independently.
  ...(isStaticExport
    ? { output: 'export', images: { unoptimized: true } }
    : {
        async rewrites() {
          const backend = process.env.LEDGERLENS_BACKEND_URL || 'http://127.0.0.1:8082';
          return [
            { source: '/api/:path*', destination: `${backend}/api/:path*` },
          ];
        },
      }),
};

export default nextConfig;
