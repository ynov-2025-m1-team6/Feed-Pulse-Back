package utils

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "Email with subdomain",
			email: "user@sub.example.com",
			want:  true,
		},
		{
			name:  "Email with plus sign",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "Empty email",
			email: "",
			want:  false,
		},
		{
			name:  "Too short email",
			email: "a@",
			want:  false,
		},
		{
			name:  "Too long email",
			email: "a@" + string(make([]byte, 252)) + ".com",
			want:  false,
		},
		{
			name:  "Email starts with @",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "Email ends with @",
			email: "user@",
			want:  false,
		},
		{
			name:  "Email with consecutive @",
			email: "user@@example.com",
			want:  false,
		},
		{
			name:  "Email with spaces",
			email: "user @example.com",
			want:  false,
		},
		{
			name:  "Invalid email format",
			email: "user@example..com",
			want:  false,
		},
		{
			name:  "Email with no @ symbol",
			email: "userexample.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
