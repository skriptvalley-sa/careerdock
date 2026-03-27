ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS cart_snapshot JSONB;

ALTER TABLE payments
    DROP CONSTRAINT IF EXISTS payments_product_type_check;

ALTER TABLE payments
    ADD CONSTRAINT payments_product_type_check
        CHECK (product_type IN (
            'starter_pack',
            'starter_refill',
            'resume_bundle',
            'ats_bundle',
            'curated_list_bundle',
            'cv_bundle',
            'cart_bundle',
            'resume_upload',
            'cv_generation',
            'rebuy_pack'
        ));
