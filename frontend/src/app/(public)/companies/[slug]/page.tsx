import type { Metadata } from 'next';
import CompanyProfile from './company-profile';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || 'https://careerdock.in';

interface CompanyMeta {
  name: string;
  slug: string;
  description?: string;
  headquarters?: string;
  hiring_status: string;
  tech_stack: string[];
  logo_url?: string;
}

async function fetchCompany(slug: string): Promise<CompanyMeta | null> {
  try {
    const res = await fetch(`${API_BASE}/api/companies/${slug}`, {
      next: { revalidate: 3600 }, // ISR: revalidate every hour
    });
    if (!res.ok) return null;
    const json = await res.json();
    return json.data as CompanyMeta;
  } catch {
    return null;
  }
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const company = await fetchCompany(slug);

  if (!company) {
    return {
      title: 'Company Not Found | CareerDock',
      description: 'The requested company could not be found.',
    };
  }

  const title = `${company.name} — Tech Profile & Career Info | CareerDock`;
  const description =
    company.description ||
    `Explore ${company.name}'s tech stack, hiring status, and career opportunities on CareerDock.`;
  const url = `${SITE_URL}/companies/${company.slug}`;

  return {
    title,
    description,
    alternates: {
      canonical: url,
    },
    openGraph: {
      title,
      description,
      url,
      siteName: 'CareerDock',
      type: 'website',
      ...(company.logo_url && {
        images: [{ url: company.logo_url, width: 200, height: 200, alt: company.name }],
      }),
    },
    twitter: {
      card: 'summary',
      title,
      description,
    },
    other: {
      'application-name': 'CareerDock',
    },
  };
}

export default function CompanyPage() {
  return <CompanyProfile />;
}
