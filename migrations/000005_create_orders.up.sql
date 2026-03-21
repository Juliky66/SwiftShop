CREATE TABLE carts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id        UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    variant_id     UUID NOT NULL,
    quantity       INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    price_snapshot NUMERIC(12,2) NOT NULL,
    UNIQUE (cart_id, variant_id)
);

CREATE INDEX idx_cart_items_cart ON cart_items (cart_id);

CREATE TABLE orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    status           TEXT NOT NULL DEFAULT 'pending',
    total_amount     NUMERIC(12,2) NOT NULL,
    delivery_address JSONB NOT NULL DEFAULT '{}',
    payment_id       UUID,
    payment_method   TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user   ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);

CREATE TABLE order_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL REFERENCES orders(id),
    variant_id   UUID NOT NULL,
    seller_id    UUID NOT NULL,
    product_name TEXT NOT NULL,
    sku          TEXT NOT NULL,
    quantity     INT NOT NULL,
    unit_price   NUMERIC(12,2) NOT NULL,
    total_price  NUMERIC(12,2) NOT NULL
);

CREATE INDEX idx_order_items_order  ON order_items (order_id);
CREATE INDEX idx_order_items_seller ON order_items (seller_id);
