package sessionManager

import (
	"testing"
	"time"
)

func TestSessionManager(t *testing.T) {
	// Initialize a new session manager with a short expiration for testing
	sm := NewSessionManager("test-secret-key", 2*time.Second)

	// Test creating a session
	t.Run("CreateSession", func(t *testing.T) {
		userUUID := "test-user-uuid"
		token, err := sm.CreateSession(userUUID)
		
		if err != nil {
			t.Errorf("CreateSession() error = %v, expected nil", err)
		}
		
		if token == "" {
			t.Error("CreateSession() returned empty token")
		}
		
		// Verify the token was added to the sessions map
		sm.mu.Lock()
		_, exists := sm.sessions[token]
		sm.mu.Unlock()
		
		if !exists {
			t.Error("Token was not added to the sessions map")
		}
	})

	// Test validating a valid session
	t.Run("ValidateSession_Valid", func(t *testing.T) {
		userUUID := "test-user-uuid-2"
		token, _ := sm.CreateSession(userUUID)
		
		valid, err := sm.ValidateSession(token)
		
		if err != nil {
			t.Errorf("ValidateSession() error = %v, expected nil", err)
		}
		
		if !valid {
			t.Error("ValidateSession() returned false for a valid token")
		}
	})

	// Test validating an invalid session
	t.Run("ValidateSession_Invalid", func(t *testing.T) {
		token := "invalid-token"
		
		valid, _ := sm.ValidateSession(token)
		
		if valid {
			t.Error("ValidateSession() returned true for an invalid token")
		}
	})

	// Test deleting a session
	t.Run("DeleteSession", func(t *testing.T) {
		userUUID := "test-user-uuid-3"
		token, _ := sm.CreateSession(userUUID)
		
		// Verify the session exists before deletion
		sm.mu.Lock()
		_, existsBefore := sm.sessions[token]
		sm.mu.Unlock()
		
		if !existsBefore {
			t.Error("Token was not added to the sessions map before deletion")
		}
		
		// Delete the session
		sm.DeleteSession(token)
		
		// Verify the session no longer exists
		sm.mu.Lock()
		_, existsAfter := sm.sessions[token]
		sm.mu.Unlock()
		
		if existsAfter {
			t.Error("Token still exists in the sessions map after deletion")
		}
	})

	// Test session expiration
	t.Run("SessionExpiration", func(t *testing.T) {
		userUUID := "test-user-uuid-4"
		token, _ := sm.CreateSession(userUUID)
		
		// Wait for the session to expire (slightly longer than expiration time)
		time.Sleep(2500 * time.Millisecond)
		
		valid, _ := sm.ValidateSession(token)
		
		if valid {
			t.Error("ValidateSession() returned true for an expired token")
		}
	})
}

func TestInitSessionManager(t *testing.T) {
	// Test initializing the global session manager instance
	InitSessionManager("test-secret-key", 1*time.Hour)
	
	if Instance == nil {
		t.Error("InitSessionManager() did not initialize the global Instance")
	}
}
