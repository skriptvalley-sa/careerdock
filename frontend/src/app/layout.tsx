import type { Metadata } from 'next';
import { Providers } from '@/components/providers';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';
import '@/styles/globals.css';

export const metadata: Metadata = {
  title: 'CareerDock — Career Intelligence for Tech Job Seekers',
  description:
    'Browse 200+ Indian tech companies, track applications, get AI-powered ATS scores, and land your dream tech job.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-white text-gray-900 antialiased">
        <Providers>
          <Header />
          <main className="min-h-[calc(100vh-4rem)]">{children}</main>
          <Footer />
        </Providers>
      </body>
    </html>
  );
}
