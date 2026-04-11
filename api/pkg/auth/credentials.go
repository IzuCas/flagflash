package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Credentials holds the current login credentials and state.
// Password is always stored as a bcrypt hash of (Salt + plainPassword).
// Salt is a UUID v4 generated once per password change. When Salt is empty
// (e.g. legacy bcrypt without salt) verification falls back to plain bcrypt.
type Credentials struct {
	Username         string    `json:"username"`
	Password         string    `json:"password"`
	Salt             string    `json:"salt,omitempty"`
	IsInitial        bool      `json:"is_initial"`
	TokensValidAfter time.Time `json:"tokens_valid_after,omitempty"`
}

func credentialsFilePath() string {
	if p := os.Getenv("CREDENTIALS_FILE"); p != "" {
		return p
	}
	return "./data/credentials.json"
}

// generateSalt returns a random UUID v4 string used as an explicit password salt.
// The UUID is stored in credentials.json alongside the bcrypt hash, so even if
// an attacker obtains the hash they must also know the UUID to brute-force it.
func generateSalt() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("auth: failed to generate salt: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // UUID version 4
	b[8] = (b[8] & 0x3f) | 0x80 // UUID variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Verify checks a plain-text password against the stored bcrypt hash.
// When a UUID salt is present it prepends salt to the plain password before
// comparison (salt+plain is what was hashed at save time).
func (c Credentials) Verify(plain string) bool {
	input := plain
	if c.Salt != "" {
		input = c.Salt + plain
	}
	if strings.HasPrefix(c.Password, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(input)) == nil
	}
	// Legacy plain-text : env var defaults that were never written to disk as bcrypt.
	return subtle.ConstantTimeCompare([]byte(plain), []byte(c.Password)) == 1
}

// HashPassword hashes (salt+plain) with bcrypt. Pass an empty salt for paths
// that do not yet use the salted scheme (transparent legacy migration).
func HashPassword(plain, salt string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(salt+plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// LoadCredentials loads stored credentials from the credentials file.
// If the file does not exist or is unreadable, it falls back to the
// AUTH_USERNAME / AUTH_PASSWORD env vars (default: guest / guest).
// Plain-text passwords (no bcrypt prefix) are auto-migrated to bcrypt+salt.
func LoadCredentials() Credentials {
	path := credentialsFilePath()
	data, err := os.ReadFile(path)
	if err == nil {
		var creds Credentials
		if json.Unmarshal(data, &creds) == nil {
			// Auto-migrate plain-text password → bcrypt + UUID salt
			if !strings.HasPrefix(creds.Password, "$2") {
				migratePlainTextPassword(&creds, path)
			}
			return creds
		}
	}

	username := os.Getenv("AUTH_USERNAME")
	if username == "" {
		username = "guest"
	}
	password := os.Getenv("AUTH_PASSWORD")
	if password == "" {
		password = "guest"
	}
	if username == "guest" || password == "guest" {
		log.Println("[SECURITY WARNING] Using default credentials (guest/guest). " +
			"Set AUTH_USERNAME and AUTH_PASSWORD environment variables before deploying to production.")
	}
	return Credentials{Username: username, Password: password, IsInitial: true}
}

// migratePlainTextPassword rehashes a plain-text password using bcrypt + a new
// UUID salt and writes the result back to disk.
// TokensValidAfter is intentionally NOT updated so existing sessions remain valid.
func migratePlainTextPassword(creds *Credentials, path string) {
	salt := generateSalt()
	hash, err := HashPassword(creds.Password, salt)
	if err != nil {
		return
	}
	creds.Password = hash
	creds.Salt = salt
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	_ = os.WriteFile(path, data, 0600)
}

// SaveCredentials generates a fresh UUID salt, hashes (salt+password) with bcrypt,
// sets TokensValidAfter to now (invalidating all existing JWTs), and persists to disk.
func SaveCredentials(username, password string) error {
	salt := generateSalt()
	hash, err := HashPassword(password, salt)
	if err != nil {
		return err
	}
	creds := Credentials{
		Username:         username,
		Password:         hash,
		Salt:             salt,
		IsInitial:        false,
		TokensValidAfter: time.Now(),
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	path := credentialsFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
