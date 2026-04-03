package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

const publicKeyHex = "3af8f9593b3331c27994f1eeacf111c727ff6015016b0af44ed3ca6934d40b13"

type Limits struct {
	MaxSites       int
	RetentionDays  int
	Funnels        bool
	Cohorts        bool
	Export         bool
	RealTime       bool
	Tier           string
}

func FreeLimits() Limits {
	return Limits{
		MaxSites: 1, RetentionDays: 30,
		Funnels: false, Cohorts: false, Export: false, RealTime: false,
		Tier: "free",
	}
}

func ProLimits() Limits {
	return Limits{
		MaxSites: 0, RetentionDays: 365,
		Funnels: true, Cohorts: true, Export: true, RealTime: true,
		Tier: "pro",
	}
}

func DefaultLimits() Limits {
	key := os.Getenv("STOCKYARD_LICENSE_KEY")
	if key == "" {
		log.Printf("[license] No license key — free tier (1 site, 30 day retention)")
		log.Printf("[license] Get Pro at https://stockyard.dev/headcount/")
		return FreeLimits()
	}
	if validateLicenseKey(key, "headcount") {
		log.Printf("[license] Valid Pro license — all features unlocked")
		return ProLimits()
	}
	log.Printf("[license] Invalid license key — running on free tier")
	return FreeLimits()
}

func LimitReached(limit, current int) bool {
	if limit == 0 { return false }
	return current >= limit
}

func validateLicenseKey(key, product string) bool {
	if !strings.HasPrefix(key, "SY-") { return false }
	key = key[3:]
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 { return false }
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil { return false }
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(sigBytes) != ed25519.SignatureSize { return false }
	pubKeyBytes, _ := hexDecode(publicKeyHex)
	if len(pubKeyBytes) != ed25519.PublicKeySize { return false }
	if !ed25519.Verify(ed25519.PublicKey(pubKeyBytes), payloadBytes, sigBytes) { return false }
	var payload struct { Product string `json:"p"`; ExpiresAt int64 `json:"x"` }
	if err := json.Unmarshal(payloadBytes, &payload); err != nil { return false }
	if payload.ExpiresAt > 0 && time.Now().Unix() > payload.ExpiresAt { return false }
	if payload.Product != "*" && payload.Product != "stockyard" && payload.Product != product { return false }
	return true
}

func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 { return nil, os.ErrInvalid }
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		h := hexVal(s[i]); l := hexVal(s[i+1])
		if h == 255 || l == 255 { return nil, os.ErrInvalid }
		b[i/2] = h<<4 | l
	}
	return b, nil
}
func hexVal(c byte) byte {
	switch {
	case c >= '0' && c <= '9': return c - '0'
	case c >= 'a' && c <= 'f': return c - 'a' + 10
	case c >= 'A' && c <= 'F': return c - 'A' + 10
	}
	return 255
}
