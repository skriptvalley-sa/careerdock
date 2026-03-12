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
