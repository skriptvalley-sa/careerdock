// API response types matching backend handler DTOs.

export interface User {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'moderator' | 'admin';
  email_verified: boolean;
  premium_since: string | null;
  current_title: string | null;
  experience_level: string | null;
  preferred_tech_stacks: string[];
  target_domains: string[];
  target_locations: string[];
  default_resume_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface DataResponse<T> {
  data: T;
}

export interface ErrorBody {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ErrorResponse {
  error: ErrorBody;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    next_cursor: string;
    has_more: boolean;
  };
}

export interface CompanyListItem {
  id: string;
  slug: string;
  name: string;
  logo_url: string | null;
  description: string | null;
  size: string | null;
  headquarters: string | null;
  tech_stack: string[];
  domains: string[];
  hiring_status: string;
  compensation_tier: string | null;
  has_rsu: boolean;
  has_rsu_refresher: boolean;
  office_modes: string[];
  updated_at: string;
}

export interface CompanyDetail extends CompanyListItem {
  founded_year: number | null;
  careers_page_url: string | null;
  glassdoor_url: string | null;
  ambitionbox_url: string | null;
  linkedin_url: string | null;
  interview_patterns: unknown;
  compensation_bands: unknown;
  last_verified_at: string | null;
  created_at: string;
}

export interface CompanyFilterParams {
  q?: string;
  size?: string;
  hiring_status?: string;
  tech_stack?: string;
  domains?: string;
  compensation_tier?: string;
  has_rsu?: string;
  headquarters?: string;
  sort?: string;
  order?: string;
  cursor?: string;
  limit?: string;
}

// --- Lists & Tracking (Sprint 2) ---

export type ApplicationStatus =
  | 'not_applied'
  | 'applied'
  | 'phone_screen'
  | 'interview'
  | 'offer'
  | 'rejected'
  | 'accepted'
  | 'withdrawn';

export type CompanyTrackingStatus =
  | 'marked'
  | 'researching'
  | 'applied'
  | 'interviewing'
  | 'offered'
  | 'accepted'
  | 'rejected';

export interface UserList {
  id: string;
  name: string;
  description: string | null;
  position: number;
  entry_count: number;
  created_at: string;
  updated_at: string;
}

export interface ListEntry {
  id: string;
  list_id: string;
  company_id: string;
  company_name: string;
  company_slug?: string;
  company_status: CompanyTrackingStatus;
  role_title: string | null;
  status: ApplicationStatus;
  date_applied: string | null;
  notes: string | null;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface ListDetail extends UserList {
  entries: ListEntry[];
}

export interface StatusHistoryItem {
  id: string;
  from_status: string | null;
  to_status: string;
  changed_at: string;
}

export interface InterviewRound {
  id: string;
  round_number: number;
  round_type: string;
  scheduled_date: string | null;
  outcome: 'passed' | 'failed' | 'pending';
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface ListCompanyFlag {
  list_id: string;
  name: string;
  contains_company: boolean;
}

export interface DashboardCounts {
  not_applied: number;
  applied: number;
  phone_screen: number;
  interview: number;
  offer: number;
  rejected: number;
  accepted: number;
  withdrawn: number;
  total: number;
}

// --- Payments & Credits (Sprint 3) ---

export interface PaymentOrder {
  payment_id: string;
  razorpay_order_id: string;
  razorpay_key_id: string;
  amount_paise: number;
  currency: string;
  product_type: string;
}

export interface PaymentRecord extends PaymentOrder {
  razorpay_payment_id?: string;
}

export interface CreditBalances {
  resume_upload: number;
  ats_check: number;
  curated_list: number;
  cv_generation: number;
}

export interface CreditTransaction {
  id: string;
  credit_type: string;
  amount: number;
  balance_after: number;
  reason: string;
  reference_id?: string;
  created_at: string;
}

// --- Resumes (Sprint 3) ---

export type ResumeStatus = 'parsing' | 'ready' | 'failed';

export interface ParsedSummary {
  years_of_experience: number;
  role_level: string;
  top_skills: string[];
  domains: string[];
}

export interface ResumeListItem {
  id: string;
  slot_number: number;
  file_name: string;
  file_size_bytes: number;
  status: ResumeStatus;
  is_default: boolean;
  ats_general_score?: number;
  parsed_data_summary?: ParsedSummary;
  created_at: string;
  updated_at: string;
}

export interface ResumeDetail {
  id: string;
  slot_number: number;
  file_name: string;
  file_size_bytes: number;
  status: ResumeStatus;
  is_default: boolean;
  parsed_data?: Record<string, unknown>;
  ats_general?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}
