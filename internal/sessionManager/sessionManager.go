package sessionManager

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var Instance *SessionManager

type SessionManager struct {
	sessions   map[string]time.Time
	mu         sync.Mutex
	secretKey  []byte
	expiration time.Duration
}

func InitSessionManager(secretKey string, expiration time.Duration) {
	Instance = NewSessionManager(secretKey, expiration)
}

func NewSessionManager(secretKey string, expiration time.Duration) *SessionManager {
	manager := &SessionManager{
		sessions:   make(map[string]time.Time),
		secretKey:  []byte(secretKey),
		expiration: expiration,
	}

	// Start a goroutine to clean up expired sessions
	go manager.cleanupExpiredSessions()

	return manager
}

func (sm *SessionManager) CreateSession(userUUID string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userUUID": userUUID,
		"exp":      time.Now().Add(sm.expiration).Unix(),
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(sm.secretKey)
	if err != nil {
		return "", err
	}

	// Store the session with its expiration time
	sm.sessions[tokenString] = time.Now().Add(sm.expiration)

	return tokenString, nil
}

func (sm *SessionManager) ValidateSession(tokenString string) (bool, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return sm.secretKey, nil
	})
	if err != nil || !token.Valid {
		return false, err
	}

	// Check if the session exists and is not expired
	expiration, exists := sm.sessions[tokenString]
	if !exists || time.Now().After(expiration) {
		return false, nil
	}

	return true, nil
}

func (sm *SessionManager) DeleteSession(tokenString string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, tokenString)
}

func (sm *SessionManager) cleanupExpiredSessions() {
	for {
		time.Sleep(time.Minute) // Run cleanup every minute

		sm.mu.Lock()
		for token, expiration := range sm.sessions {
			if time.Now().After(expiration) {
				delete(sm.sessions, token)
			}
		}
		sm.mu.Unlock()
	}
}

// GetSecretKey returns the secret key used for signing JWT tokens
func (sm *SessionManager) GetSecretKey() []byte {
	return sm.secretKey
}
