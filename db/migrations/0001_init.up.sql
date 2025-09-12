CREATE TABLE orders (
  id         TEXT        PRIMARY KEY,
  amount_jpy BIGINT      NOT NULL CHECK (amount_jpy > 0),
  status     TEXT        NOT NULL CHECK (status IN ('PENDING','PAID','CANCELED')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payments (
  id             TEXT        PRIMARY KEY,
  order_id       TEXT        NOT NULL REFERENCES orders(id),
  method         TEXT        NOT NULL,
  provider       TEXT        NOT NULL,
  provider_tx_id TEXT        NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_payments_provider_tx UNIQUE (provider, provider_tx_id)
);

CREATE TABLE payment_events (
  id         TEXT        PRIMARY KEY,
  order_id   TEXT        NOT NULL REFERENCES orders(id),
  type       TEXT        NOT NULL,
  payload    JSONB       NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payment_events_order_created ON payment_events(order_id, created_at DESC);
