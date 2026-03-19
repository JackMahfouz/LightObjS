CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    api_secret TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_clients_user_id ON clients (user_id);
CREATE INDEX idx_clients_api_key ON clients (api_key);