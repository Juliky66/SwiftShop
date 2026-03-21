CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  UUID REFERENCES categories(id),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_parent ON categories (parent_id);
CREATE INDEX idx_categories_slug   ON categories (slug);

CREATE TABLE products (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id    UUID NOT NULL REFERENCES sellers(id),
    category_id  UUID NOT NULL REFERENCES categories(id),
    name         TEXT NOT NULL,
    slug         TEXT NOT NULL UNIQUE,
    description  TEXT,
    brand        TEXT,
    status       TEXT NOT NULL DEFAULT 'draft',
    rating       NUMERIC(3,2) DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_category ON products (category_id);
CREATE INDEX idx_products_seller   ON products (seller_id);
CREATE INDEX idx_products_status   ON products (status);
CREATE INDEX idx_products_fts ON products
    USING GIN (to_tsvector('russian', name || ' ' || coalesce(description, '') || ' ' || coalesce(brand, '')));

CREATE TABLE product_variants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku        TEXT NOT NULL UNIQUE,
    price      NUMERIC(12,2) NOT NULL,
    old_price  NUMERIC(12,2),
    stock      INT NOT NULL DEFAULT 0,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_variants_product ON product_variants (product_id);
CREATE INDEX idx_variants_sku     ON product_variants (sku);

CREATE TABLE product_images (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_images_product ON product_images (product_id);

CREATE TABLE attributes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id    UUID NOT NULL REFERENCES categories(id),
    name           TEXT NOT NULL,
    attribute_type TEXT NOT NULL
);
