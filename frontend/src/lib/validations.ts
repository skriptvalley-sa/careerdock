import { z } from 'zod';

/**
 * Password validation matching the backend policy:
 * min 8, max 72 (bcrypt limit), 1 upper + 1 lower + 1 digit.
 */
export const passwordSchema = z
  .string()
  .min(8, 'Password must be at least 8 characters')
  .max(72, 'Password must be at most 72 characters')
  .refine((val) => /[A-Z]/.test(val), 'Must contain at least 1 uppercase letter')
  .refine((val) => /[a-z]/.test(val), 'Must contain at least 1 lowercase letter')
  .refine((val) => /[0-9]/.test(val), 'Must contain at least 1 digit');

export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email').max(255),
  password: z.string().min(1, 'Password is required'),
});

export const registerSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  email: z.string().email('Please enter a valid email').max(255),
  password: passwordSchema,
});

export const forgotPasswordSchema = z.object({
  email: z.string().email('Please enter a valid email').max(255),
});

export const resetPasswordSchema = z.object({
  password: passwordSchema,
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>;
