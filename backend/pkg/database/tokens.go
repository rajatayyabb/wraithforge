package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashToken returns the hex SHA-256 of a token. Only the hash is ever stored,
// so a database leak does not expose usable tokens.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GenerateToken returns a new random 256-bit token as a hex string, suitable
// for use as a Bearer token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// CreateAPIToken generates a token, stores its hash under the given label, and
// returns the plaintext token. The plaintext is returned only here and is not
// recoverable afterwards.
func (db *DB) CreateAPIToken(ctx context.Context, label string) (token, id string, err error) {
	token, err = GenerateToken()
	if err != nil {
		return "", "", err
	}
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO api_tokens (label, token_hash) VALUES ($1, $2) RETURNING id`,
		label, HashToken(token),
	).Scan(&id)
	if err != nil {
		return "", "", err
	}
	return token, id, nil
}

// ValidateAPIToken reports whether a presented token matches a non-revoked
// token in the database. The lookup is by hash, so the comparison is performed
// over fixed-width hashes rather than the raw secret.
func (db *DB) ValidateAPIToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	var exists bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS (
		    SELECT 1 FROM api_tokens
		    WHERE token_hash = $1 AND revoked_at IS NULL
		 )`,
		HashToken(token),
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// RevokeAPIToken marks a token (identified by its id) revoked.
func (db *DB) RevokeAPIToken(ctx context.Context, id string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE api_tokens SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}
