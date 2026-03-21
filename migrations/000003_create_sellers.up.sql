CREATE TABLE sellers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL UNIQUE REFERENCES users(id),
    brand_name TEXT NOT NULL,
    inn        TEXT NOT NULL UNIQUE,
    status     TEXT NOT NULL DEFAULT 'pending',
    rating     NUMERIC(3,2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sellers_status ON sellers (status);
