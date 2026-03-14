import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // Output standalone build for Docker deployment
  output: 'standalone',

  // Disable the floating "N" dev toolbar in development
  devIndicators: false,
};

export default nextConfig;
