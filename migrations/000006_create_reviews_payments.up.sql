CREATE TABLE reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL,
    user_id     UUID NOT NULL,
    order_id    UUID,
    rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title       TEXT,
    body        TEXT,
    is_approved BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, user_id)
);

CREATE INDEX idx_reviews_product ON reviews (product_id);
CREATE INDEX idx_reviews_user    ON reviews (user_id);

CREATE TABLE review_media (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id  UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    media_type TEXT NOT NULL
);

CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL UNIQUE,
    provider    TEXT NOT NULL,
    provider_id TEXT,
    status      TEXT NOT NULL DEFAULT 'created',
    amount      NUMERIC(12,2) NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'RUB',
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order  ON payments (order_id);
CREATE INDEX idx_payments_status ON payments (status);
