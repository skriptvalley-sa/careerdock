import type { Metadata } from 'next';
import { Providers } from '@/components/providers';
import { Header } from '@/components/layout/header';
import { AppShell } from '@/components/layout/app-shell';
import '@/styles/globals.css';

export const metadata: Metadata = {
  title: 'CareerDock — Career Intelligence for Tech Job Seekers',
  description:
    'Browse 200+ Indian tech companies, track applications, get AI-powered ATS scores, and land your dream tech job.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen overscroll-none bg-surface text-[var(--color-text)] antialiased">
        <Providers>
          <Header />
          {/* pt-14 compensates for the fixed header (h-14 = 56px) */}
          <div className="pt-14">
            <AppShell>{children}</AppShell>
          </div>
        </Providers>
      </body>
    </html>
  );
}
