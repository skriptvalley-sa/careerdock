-- Seed default feature flags for payment/premium gating.
INSERT INTO feature_flags (key, enabled, description) VALUES
  ('payments_enabled', FALSE, 'Master switch for the Razorpay payment flow. When disabled, the purchase UI is hidden.'),
  ('premium_bypass',   FALSE, 'When enabled, all users are treated as premium (useful during beta/testing).')
ON CONFLICT (key) DO NOTHING;
