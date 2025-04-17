package utils

func IsValidEmail(email string) bool {
	// Simple email validation logic
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	if email[0] == '@' || email[len(email)-1] == '@' {
		return false
	}
	for i := 0; i < len(email)-1; i++ {
		if email[i] == '@' && email[i+1] == '@' {
			return false
		}
	}
	return true
}
