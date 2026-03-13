import Link from 'next/link';
import { Search, FileText, Target, BarChart3 } from 'lucide-react';

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
  return (
    <div>
      {/* Hero */}
      <section className="mx-auto max-w-7xl px-4 py-20 text-center sm:px-6 lg:px-8">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
          Career Intelligence for
          <span className="text-blue-600"> Indian Tech Professionals</span>
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-600">
          Research companies, analyze your resume with AI, get ATS scores, and track applications
          &mdash; all in one place. Built specifically for the Indian tech job market.
        </p>
        <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
          <Link
            href="/companies"
            className="rounded-lg bg-blue-600 px-6 py-3 text-sm font-semibold text-white shadow hover:bg-blue-700"
          >
            Browse Companies
          </Link>
          <Link
            href="/register"
            className="rounded-lg border border-gray-300 px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
          >
            Create Free Account
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="border-t border-gray-100 bg-gray-50 py-20">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <h2 className="text-center text-2xl font-bold text-gray-900">
            Everything you need to land your next role
          </h2>
          <div className="mt-12 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((f) => (
              <div key={f.title} className="rounded-lg bg-white p-6 shadow-sm">
                <f.icon className="h-8 w-8 text-blue-600" />
                <h3 className="mt-4 text-base font-semibold text-gray-900">{f.title}</h3>
                <p className="mt-2 text-sm text-gray-600">{f.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20">
        <div className="mx-auto max-w-3xl px-4 text-center sm:px-6 lg:px-8">
          <h2 className="text-2xl font-bold text-gray-900">Start with the free tier</h2>
          <p className="mt-4 text-gray-600">
            Browse the full company directory, create up to 3 company lists, and track your
            applications &mdash; completely free. Upgrade when you&apos;re ready for AI-powered
            insights.
          </p>
          <div className="mt-8">
            <Link
              href="/register"
              className="rounded-lg bg-blue-600 px-8 py-3 text-sm font-semibold text-white shadow hover:bg-blue-700"
            >
              Get Started Free
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
