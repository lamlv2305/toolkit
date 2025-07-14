package fat

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
)

type Signature struct {
	whitelisted       []string
	secrets           []string
	randomPositions   []int
	timeWindowSeconds int64
	hexChars          string
}

type SignatureOption func(*Signature)

func WithWhitelisted(whitelisted []string) SignatureOption {
	return func(s *Signature) {
		s.whitelisted = whitelisted
	}
}

func WithSecrets(secrets []string) SignatureOption {
	return func(s *Signature) {
		s.secrets = secrets
	}
}

func WithRandomPositions(positions []int) SignatureOption {
	return func(s *Signature) {
		s.randomPositions = positions
	}
}

func WithTimeWindowSeconds(seconds int64) SignatureOption {
	return func(s *Signature) {
		s.timeWindowSeconds = seconds
	}
}

// NewSignature creates a new Signature instance with optional configurations
func NewSignature(options ...SignatureOption) *Signature {
	sig := &Signature{
		timeWindowSeconds: 10,
		hexChars:          "0123456789abcdef",
	}

	for _, opt := range options {
		opt(sig)
	}

	if sig.whitelisted == nil {
		sig.whitelisted = []string{}
	}
	if sig.secrets == nil {
		sig.secrets = []string{}
	}
	if sig.randomPositions == nil {
		sig.randomPositions = []int{}
	}

	return sig
}

// generateRandomChar generates a random hex character
func (s *Signature) generateRandomChar() string {
	return string(s.hexChars[rand.Intn(len(s.hexChars))])
}

// insertRandomAtPosition inserts random characters at specific positions
func (s *Signature) insertRandomAtPosition(sig string, positions []int) string {
	if len(positions) == 0 {
		return sig
	}

	result := []rune(sig)

	// Sort positions in descending order to insert from right to left
	// This prevents position shifts when inserting
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		if pos > len(result) {
			continue
		}

		randomChar := []rune(s.generateRandomChar())
		result = append(result[:pos], append(randomChar, result[pos:]...)...)
	}

	return string(result)
}

// removeRandomAtPosition removes characters at specific positions
func (s *Signature) removeRandomAtPosition(sig string, positions []int) string {
	if len(positions) == 0 {
		return sig
	}

	result := []rune(sig)

	// Sort positions in descending order to remove from right to left
	// This prevents position shifts when removing
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		if pos >= len(result) {
			continue
		}

		result = append(result[:pos], result[pos+1:]...)
	}

	return string(result)
}

// getCurrentTimeWindows returns current and previous time windows for validation
func (s *Signature) getCurrentTimeWindows() []int64 {
	now := time.Now().Unix()
	currentWindow := (now / s.timeWindowSeconds) * s.timeWindowSeconds
	return []int64{currentWindow, currentWindow - s.timeWindowSeconds}
}

// generateSignatureHash creates an MD5 hash for bundleID and secret
func (s *Signature) generateSignatureHash(bundleID, secret string) string {
	data := bundleID + secret
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// isValidSignatureForBundle checks if signature matches any time window for a bundle
func (s *Signature) isValidSignatureForBundle(cleanSig, bundleID string) bool {
	timeWindows := s.getCurrentTimeWindows()

	for _, timeWindow := range timeWindows {
		secretIndex := timeWindow % int64(len(s.secrets))
		secret := s.secrets[secretIndex]
		expectedSig := s.generateSignatureHash(bundleID, secret)

		if cleanSig == expectedSig {
			return true
		}
	}

	return false
}

func (s *Signature) Validate(sig string) bool {
	if len(s.secrets) == 0 || len(s.whitelisted) == 0 {
		return false
	}

	cleanSig := s.removeRandomAtPosition(sig, s.randomPositions)

	for _, bundleID := range s.whitelisted {
		if s.isValidSignatureForBundle(cleanSig, bundleID) {
			return true
		}
	}

	return false
}

// Generate generates a signature for a given bundle ID using current time
// This is a helper function for testing/client implementation
func (s *Signature) Generate(bundleID string) string {
	if len(s.secrets) == 0 {
		return ""
	}

	now := time.Now().Unix()
	currentWindow := (now / s.timeWindowSeconds) * s.timeWindowSeconds
	secretIndex := currentWindow % int64(len(s.secrets))
	secret := s.secrets[secretIndex]

	baseSig := s.generateSignatureHash(bundleID, secret)
	return s.insertRandomAtPosition(baseSig, s.randomPositions)
}
