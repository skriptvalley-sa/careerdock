export const SUPPORT_EMAIL = 'support@skirptvalley.com';

export type CreditTypeKey =
  | 'resume_upload'
  | 'ats_check'
  | 'curated_list'
  | 'cv_generation';

export type ProductType =
  | 'starter_pack'
  | 'starter_refill'
  | 'resume_bundle'
  | 'ats_bundle'
  | 'curated_list_bundle'
  | 'cv_bundle';

export type PremiumProductType = Exclude<ProductType, 'starter_pack'>;

export interface ProductDefinition {
  productType: ProductType;
  name: string;
  price: number;
  description: string;
  premiumOnly: boolean;
  available: boolean;
  statusLabel?: string;
  credits: Partial<Record<CreditTypeKey, number>>;
}

export interface PremiumProductDefinition extends ProductDefinition {
  productType: PremiumProductType;
  premiumOnly: true;
}

const creditLabels: Record<CreditTypeKey, string> = {
  resume_upload: 'Resume',
  ats_check: 'ATS',
  curated_list: 'Lists',
  cv_generation: 'Cover Letter',
};

const creditUnitLabels: Record<CreditTypeKey, string> = {
  resume_upload: 'uploads',
  ats_check: 'checks',
  curated_list: 'lists',
  cv_generation: 'cover letters',
};

export const starterPackProduct: ProductDefinition = {
  productType: 'starter_pack',
  name: 'Starter Pack',
  price: 449,
  description: 'One-time activation pack for premium features and AI credits.',
  premiumOnly: false,
  available: true,
  credits: {
    resume_upload: 10,
    ats_check: 50,
    curated_list: 10,
    cv_generation: 50,
  },
};

export const premiumShopProducts: PremiumProductDefinition[] = [
  {
    productType: 'starter_refill',
    name: 'Starter Refill Pack',
    price: 399,
    description: 'Full premium credit refill for existing premium users.',
    premiumOnly: true,
    available: true,
    credits: starterPackProduct.credits,
  },
  {
    productType: 'resume_bundle',
    name: 'Resume Bundle',
    price: 89,
    description: 'Top up resume upload credits.',
    premiumOnly: true,
    available: true,
    credits: { resume_upload: 10 },
  },
  {
    productType: 'ats_bundle',
    name: 'ATS Bundle',
    price: 229,
    description: 'Refill ATS checks for company, job, and resume scoring.',
    premiumOnly: true,
    available: true,
    credits: { ats_check: 50 },
  },
  {
    productType: 'curated_list_bundle',
    name: 'Curated Lists Bundle',
    price: 59,
    description: 'Generate more AI-ranked company lists.',
    premiumOnly: true,
    available: true,
    credits: { curated_list: 5 },
  },
  {
    productType: 'cv_bundle',
    name: 'Cover Letter Bundle',
    price: 0,
    description: 'Generate tailored cover letters from your resume, target company, and job description.',
    premiumOnly: true,
    available: false,
    statusLabel: 'Coming soon',
    credits: { cv_generation: 50 },
  },
];

export const productCatalog: Record<ProductType, ProductDefinition> = {
  starter_pack: starterPackProduct,
  starter_refill: premiumShopProducts[0],
  resume_bundle: premiumShopProducts[1],
  ats_bundle: premiumShopProducts[2],
  curated_list_bundle: premiumShopProducts[3],
  cv_bundle: premiumShopProducts[4],
};

export const creditUpsellProducts: Record<
  CreditTypeKey,
  {
    name: string;
    productType: PremiumProductType;
    price: number;
    quantity: number;
    unitLabel: string;
    available?: boolean;
  }
> = {
  resume_upload: {
    name: 'Resume Bundle',
    productType: 'resume_bundle',
    price: 89,
    quantity: 10,
    unitLabel: creditUnitLabels.resume_upload,
  },
  ats_check: {
    name: 'ATS Bundle',
    productType: 'ats_bundle',
    price: 229,
    quantity: 50,
    unitLabel: creditUnitLabels.ats_check,
  },
  curated_list: {
    name: 'Curated Lists Bundle',
    productType: 'curated_list_bundle',
    price: 59,
    quantity: 5,
    unitLabel: creditUnitLabels.curated_list,
  },
  cv_generation: {
    name: 'Cover Letter Bundle',
    productType: 'cv_bundle',
    price: 0,
    quantity: 50,
    unitLabel: creditUnitLabels.cv_generation,
    available: false,
  },
};

export function formatCreditSummary(
  credits: Partial<Record<CreditTypeKey, number>>,
): string {
  return (Object.entries(credits) as Array<[CreditTypeKey, number]>)
    .map(([creditType, amount]) => `${creditLabels[creditType]} x${amount}`)
    .join(' · ');
}

export function formatCreditLineItems(
  credits: Partial<Record<CreditTypeKey, number>>,
): string[] {
  return (Object.entries(credits) as Array<[CreditTypeKey, number]>).map(
    ([creditType, amount]) => `${amount} ${creditUnitLabels[creditType]}`,
  );
}

export function getCreditPurchaseHref(isPremium: boolean): string {
  return isPremium ? '/shop' : '/pricing';
}
