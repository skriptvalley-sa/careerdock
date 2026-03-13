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
