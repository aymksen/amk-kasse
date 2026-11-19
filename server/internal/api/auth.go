package api

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookie = "amk_admin"
	sessionIdle   = 30 * time.Minute
	maxFailures   = 5
	lockDuration  = time.Minute
	pbkdfRounds   = 200_000
)

// auth keeps admin sessions in memory: a restart simply asks for the PIN again.
type auth struct {
	mu          sync.Mutex
	sessions    map[string]time.Time
	failures    int
	lockedUntil time.Time
	now         func() time.Time
}

func newAuth(now func() time.Time) *auth {
	return &auth{sessions: map[string]time.Time{}, now: now}
}

func (a *auth) locked() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.now().Before(a.lockedUntil)
}

func (a *auth) recordFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures++
	if a.failures >= maxFailures {
		a.failures = 0
		a.lockedUntil = a.now().Add(lockDuration)
	}
}

func (a *auth) startSession() string {
	token := randomHex(32)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures = 0
	a.sessions[token] = a.now().Add(sessionIdle)
	return token
}

// valid checks a token and extends the session, so it only expires when idle.
func (a *auth) valid(token string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	expires, ok := a.sessions[token]
	if !ok || a.now().After(expires) {
		delete(a.sessions, token)
		return false
	}
	a.sessions[token] = a.now().Add(sessionIdle)
	return true
}

func (a *auth) end(token string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, token)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // the system random source failing is not recoverable
	}
	return hex.EncodeToString(b)
}

func hashPIN(pin string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pin, salt, pbkdfRounds, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(key), nil
}

func checkPIN(pin, stored string) bool {
	saltHex, keyHex, ok := strings.Cut(stored, "$")
	if !ok {
		return false
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(keyHex)
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pin, salt, pbkdfRounds, 32)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}
