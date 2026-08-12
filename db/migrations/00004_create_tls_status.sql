-- +goose Up
CREATE TABLE tls_status(
    "service_name" TEXT PRIMARY KEY,
    "fingerprint" TEXT NOT NULL, -- sha256 of the soonest certs DER
    "not_after" INTEGER NOT NULL, -- soonest expiry in chain
    "subject" TEXT,
    "issuer" TEXT,
    "is_expired" INTEGER NOT NULL DEFAULT 0,
    "chain" TEXT, -- full parsed chain as json
    "first_seen" TEXT NOT NULL,
    "last_checked" TEXT NOT NULL
);

-- +goose Down
DROP TABLE tls_status;
