'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Search, FileText, Target, BarChart3 } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';

const features = [
  {
    icon: Search,
    title: 'Company Directory',
    description:
      'Browse 200+ Indian tech companies. Filter by size, tech stack, compensation tier, and hiring status.',
  },
  {
    icon: FileText,
    title: 'AI Resume Analysis',
    description:
      'Upload your resume and get instant AI-powered feedback, skill extraction, and improvement suggestions.',
  },
  {
    icon: Target,
    title: 'ATS Scoring',
    description:
      'Get general, company-specific, and job-specific ATS scores to optimize your applications.',
  },
  {
    icon: BarChart3,
    title: 'Application Tracker',
    description:
      'Track your job applications across companies with status updates, notes, and timeline views.',
  },
];

export default function Home() {
  const { isAuthenticated, isLoading } = useAuthStore();
  const router = useRouter();

  // Redirect authenticated users to dashboard
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace('/dashboard');
    }
  }, [isLoading, isAuthenticated, router]);

  // Show nothing while loading or redirecting
  if (isLoading || isAuthenticated) {
    return null;
  }

  return (
    <div>
      {/* Hero */}
      <section className="mx-auto max-w-7xl px-4 py-20 text-center sm:px-6 lg:px-8">
        <h1 className="text-4xl font-bold tracking-tight text-slate-100 sm:text-5xl">
          Career Intelligence for
          <span className="text-[#00f0ff] text-glow-cyan"> Indian Tech Professionals</span>
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-slate-400">
          Research companies, analyze your resume with AI, get ATS scores, and track applications
          &mdash; all in one place. Built specifically for the Indian tech job market.
        </p>
        <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
          <Link
            href="/companies"
            className="btn-neon rounded-lg px-6 py-3 text-sm font-semibold shadow"
          >
            Browse Companies
          </Link>
          <Link
            href="/register"
            className="rounded-lg border border-[#00f0ff]/20 px-6 py-3 text-sm font-semibold text-[#00f0ff] hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 transition-all"
          >
            Create Free Account
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="border-t border-edge bg-overlay py-20">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <h2 className="text-center text-2xl font-bold text-slate-100">
            Everything you need to land your next role
          </h2>
          <div className="mt-12 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((f) => (
              <div key={f.title} className="card-neon-hover rounded-lg bg-card p-6 shadow-sm border border-edge">
                <f.icon className="h-8 w-8 text-[#00f0ff]" />
                <h3 className="mt-4 text-base font-semibold text-slate-100">{f.title}</h3>
                <p className="mt-2 text-sm text-slate-400">{f.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20">
        <div className="mx-auto max-w-3xl px-4 text-center sm:px-6 lg:px-8">
          <h2 className="text-2xl font-bold text-slate-100">Start with the free tier</h2>
          <p className="mt-4 text-slate-400">
            Browse the full company directory, create up to 3 company lists, and track your
            applications &mdash; completely free. Upgrade when you&apos;re ready for AI-powered
            insights.
          </p>
          <div className="mt-8">
            <Link
              href="/register"
              className="btn-neon rounded-lg px-8 py-3 text-sm font-semibold shadow"
            >
              Get Started Free
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
