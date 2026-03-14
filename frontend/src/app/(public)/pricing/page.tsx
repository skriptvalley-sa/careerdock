import Link from 'next/link';
import { Check } from 'lucide-react';

const freeTier = {
  name: 'Free',
  price: '0',
  description: 'Get started with the essentials',
  features: [
    'Full company directory access',
    'Up to 3 custom company lists',
    'Application tracking (status, dates, notes)',
    'Basic search & filters',
  ],
  cta: 'Create Free Account',
  href: '/register',
  highlight: false,
};

const starterPack = {
  name: 'Starter Pack',
  price: '299',
  description: 'One-time purchase — no subscription',
  features: [
    'Everything in Free',
    'Upload up to 3 resumes (PDF)',
    'AI resume analysis & skill extraction',
    'General ATS score',
    'Company-specific ATS score',
    'Job-specific ATS score',
    'AI-curated company matching',
  ],
  cta: 'Get Starter Pack',
  href: '/register',
  highlight: true,
};

const creditPacks = [
  { credits: 5, price: '99' },
  { credits: 15, price: '249' },
  { credits: 50, price: '699' },
];

function PricingCard({
  plan,
}: {
  plan: typeof freeTier;
}) {
  return (
    <div
      className={`flex flex-col rounded-xl border p-8 ${
        plan.highlight
          ? 'border-[#00f0ff]/40 shadow-lg shadow-[#00f0ff]/5 ring-1 ring-[#00f0ff]/30 glow-cyan'
          : 'border-edge card-neon-hover'
      }`}
    >
      {plan.highlight && (
        <span className="-mt-12 mb-4 self-start rounded-full bg-[#00f0ff]/15 border border-[#00f0ff]/30 px-3 py-1 text-xs font-semibold text-[#00f0ff]">
          Most Popular
        </span>
      )}
      <h3 className="text-lg font-semibold text-slate-100">{plan.name}</h3>
      <p className="mt-1 text-sm text-slate-500">{plan.description}</p>
      <div className="mt-6">
        <span className="text-4xl font-bold text-slate-100">{plan.price === '0' ? 'Free' : `₹${plan.price}`}</span>
        {plan.price !== '0' && <span className="text-sm text-slate-500"> one-time</span>}
      </div>
      <ul className="mt-8 flex-1 space-y-3">
        {plan.features.map((f) => (
          <li key={f} className="flex items-start gap-2 text-sm text-slate-300">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-[#00f0ff]" />
            {f}
          </li>
        ))}
      </ul>
      <Link
        href={plan.href}
        className={`mt-8 block rounded-lg px-4 py-2.5 text-center text-sm font-semibold ${
          plan.highlight
            ? 'btn-neon'
            : 'border border-[#00f0ff]/20 text-[#00f0ff] hover:bg-[#00f0ff]/5 hover:border-[#00f0ff]/40 transition-all'
        }`}
      >
        {plan.cta}
      </Link>
    </div>
  );
}

export default function PricingPage() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-slate-100">Simple, transparent pricing</h1>
        <p className="mt-4 text-lg text-slate-400">
          Start free. Pay once for AI features. No subscriptions.
        </p>
      </div>

      {/* Plans */}
      <div className="mt-12 grid gap-8 sm:grid-cols-2">
        <PricingCard plan={freeTier} />
        <PricingCard plan={starterPack} />
      </div>

      {/* Credit Packs */}
      <section className="mt-16">
        <h2 className="text-center text-xl font-bold text-slate-100">Need more credits?</h2>
        <p className="mt-2 text-center text-sm text-slate-400">
          Buy additional AI credits a la carte after your Starter Pack.
        </p>
        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          {creditPacks.map((pack) => (
            <div
              key={pack.credits}
              className="card-neon-hover rounded-lg border border-edge p-6 text-center"
            >
              <div className="text-2xl font-bold text-slate-100">{pack.credits} credits</div>
              <div className="mt-1 text-lg font-semibold text-[#00f0ff]">₹{pack.price}</div>
              <div className="mt-1 text-xs text-slate-500">
                ₹{Math.round(Number(pack.price) / pack.credits)}/credit
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* FAQ hint */}
      <section className="mt-16 text-center">
        <h2 className="text-xl font-bold text-slate-100">Questions?</h2>
        <p className="mt-2 text-sm text-slate-400">
          Reach out at{' '}
          <span className="font-medium text-slate-100">support@careerdock.in</span> and
          we&apos;ll get back to you within 24 hours.
        </p>
      </section>
    </div>
  );
}
