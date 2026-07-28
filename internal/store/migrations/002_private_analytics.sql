ALTER TABLE dogelytics_request_logs
    ALTER COLUMN client_ip DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS client_fingerprint BYTEA,
    ADD COLUMN IF NOT EXISTS wallet_fingerprint BYTEA;

CREATE INDEX IF NOT EXISTS idx_dgl_request_logs_wallet_fingerprint
    ON dogelytics_request_logs(wallet_fingerprint)
    WHERE wallet_fingerprint IS NOT NULL;

CREATE TABLE IF NOT EXISTS dogelytics_analytics_totals (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    successful_requests BIGINT NOT NULL DEFAULT 0
);

INSERT INTO dogelytics_analytics_totals (id, successful_requests)
VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS dogelytics_analytics_wallets (
    wallet_fingerprint BYTEA PRIMARY KEY,
    first_seen TIMESTAMPTZ NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS dogelytics_analytics_hourly (
    bucket TIMESTAMPTZ NOT NULL,
    api_key TEXT NOT NULL DEFAULT '',
    successful_requests BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket, api_key)
);

CREATE TABLE IF NOT EXISTS dogelytics_analytics_state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
