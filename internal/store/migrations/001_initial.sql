CREATE TABLE IF NOT EXISTS dogelytics_users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dogelytics_api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES dogelytics_users(id) ON DELETE CASCADE,
    kid TEXT UNIQUE NOT NULL,
    secret_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_dgl_apikeys_user_id ON dogelytics_api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_dgl_apikeys_expires_at ON dogelytics_api_keys(expires_at);

CREATE TABLE IF NOT EXISTS dogelytics_request_logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_ip TEXT NOT NULL,
    api_key TEXT,
    wallet_address TEXT,
    success BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_dgl_request_logs_timestamp ON dogelytics_request_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_dgl_request_logs_api_key ON dogelytics_request_logs(api_key);

CREATE TABLE IF NOT EXISTS dogelytics_conversion_rates (
    currency TEXT PRIMARY KEY,
    rate NUMERIC NOT NULL,
    coingecko_updated_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
