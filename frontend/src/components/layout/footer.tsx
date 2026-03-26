import Link from 'next/link';

export function Footer() {
  return (
    <footer className="border-t border-edge bg-overlay">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Product</h3>
            <ul className="mt-4 space-y-2">
              <li>
                <Link href="/companies" className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
                  Company Directory
                </Link>
              </li>
              <li>
                <Link href="/pricing" className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
                  Pricing
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Features</h3>
            <ul className="mt-4 space-y-2">
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">ATS Scoring</span>
              </li>
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">Company Lists</span>
              </li>
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">Application Tracking</span>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Resources</h3>
            <ul className="mt-4 space-y-2">
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">Blog (coming soon)</span>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Legal</h3>
            <ul className="mt-4 space-y-2">
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">Privacy Policy</span>
              </li>
              <li>
                <span className="text-sm text-[var(--color-text-muted)]">Terms of Service</span>
              </li>
            </ul>
          </div>
        </div>
        <div className="mt-8 border-t border-edge pt-8 text-center">
          <p className="text-sm text-[var(--color-text-muted)]">
            &copy; {new Date().getFullYear()} CareerDock by SkriptValley. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
