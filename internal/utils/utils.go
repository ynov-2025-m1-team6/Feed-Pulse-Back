package utils

import (
	"regexp"
	"strings"
)

func IsValidEmail(email string) bool {
	// Check for empty or invalid length
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	
	// Check for spaces
	if strings.Contains(email, " ") {
		return false
	}
	
	// Check for @ symbol
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return false
	}
	
	// Basic positioning rules
	if atIndex == 0 || atIndex == len(email)-1 {
		return false
	}
	
	// Check for consecutive @ symbols
	if strings.Contains(email, "@@") {
		return false
	}
	
	// Check domain part
	domain := email[atIndex+1:]
	if strings.Contains(domain, "..") {
		return false
	}
	
	// Use regex for a more comprehensive check
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
