import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Output standalone build for Docker deployment
  output: 'standalone',

  // Disable the floating "N" dev toolbar in development
  devIndicators: false,

  // Allow the public dev hostname when running behind Nginx on the VPS.
  // Without this, Next.js dev mode emits cross-origin warnings for assets
  // served through the proxied hostname.
  allowedDevOrigins: ['careerdock.skriptvalley.com'],
};

export default nextConfig;
