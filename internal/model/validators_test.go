package model

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"", false},
		{"notanemail", false},
		{"@domain.com", false},
		{"user@", false},
	}
	for _, tt := range tests {
		r := ValidateEmail(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateEmail(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateCreditCard(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"4242424242424242", true},
		{"4242 4242 4242 4242", true},
		{"4242-4242-4242-4242", true},
		{"5555555555554444", true},
		{"378282246310005", true},
		{"6011111111111117", true},
		{"1234567890123456", false},
		{"", false},
		{"abcd", false},
	}
	for _, tt := range tests {
		r := ValidateCreditCard(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateCreditCard(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateISBN(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"0-306-40615-2", true},
		{"0306406152", true},
		{"978-0-306-40615-7", true},
		{"9780306406157", true},
		{"0-306-40615-3", false},
		{"", false},
		{"123", false},
	}
	for _, tt := range tests {
		r := ValidateISBN(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateISBN(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateIBAN(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"GB82 WEST 1234 56 9876 5432", true},
		{"GB82WEST12345698765432", true},
		{"DE89 3704 0044 0532 0130 00", true},
		{"FR1420041010050500013M02606", true},
		{"GB82 WEST 1234 56 9876 5433", false},
		{"", false},
		{"123", false},
	}
	for _, tt := range tests {
		r := ValidateIBAN(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateIBAN(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateBIC(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"DEUTDEFF", true},
		{"DEUTDEFF500", true},
		{"SBICKEN1", true},
		{"", false},
		{"ABC", false},
		{"12345678", false},
	}
	for _, tt := range tests {
		r := ValidateBIC(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateBIC(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateVIN(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1HGBH41JXMN109186", true},
		{"JT3AC94R540123456", true},
		{"", false},
		{"1HGBH41JXMN109187", false},
		{"1HGBH41JXMN10918", false},
		{"1HGBH41JXMN1091866", false},
	}
	for _, tt := range tests {
		r := ValidateVIN(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateVIN(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateSSN(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123-45-6789", false},
		{"001-01-0001", true},
		{"123456789", false},
		{"666-12-3456", false},
		{"900-12-3456", false},
		{"123-01-0001", true},
		{"", false},
		{"123-45-678", false},
	}
	for _, tt := range tests {
		r := ValidateSSN(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateSSN(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"example.co.uk", true},
		{"", false},
		{"-example.com", false},
		{"example", false},
		{"a.b", false},
	}
	for _, tt := range tests {
		r := ValidateDomain(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateDomain(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateMAC(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"01:23:45:67:89:ab", true},
		{"01-23-45-67-89-ab", true},
		{"0123.4567.89ab", true},
		{"0123456789ab", true},
		{"", false},
		{"01:23:45:67:89", false},
		{"zz:zz:zz:zz:zz:zz", false},
	}
	for _, tt := range tests {
		r := ValidateMAC(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateMAC(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidatePostal(t *testing.T) {
	tests := []struct {
		input   string
		country string
		valid   bool
	}{
		{"10001", "US", true},
		{"10001-1234", "US", true},
		{"SW1A 1AA", "UK", true},
		{"K1A 0B1", "CA", true},
		{"10115", "DE", true},
		{"75001", "FR", true},
		{"2000", "AU", true},
		{"100-0001", "JP", true},
		{"01310-100", "BR", true},
		{"110001", "IN", true},
		{"99999", "US", true},
		{"9999", "US", false},
		{"", "US", false},
		{"10001", "XX", false},
	}
	for _, tt := range tests {
		r := ValidatePostal(tt.input, tt.country)
		if r.Valid != tt.valid {
			t.Errorf("ValidatePostal(%q, %q) = %v, want %v, detail: %s", tt.input, tt.country, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"#fff", true},
		{"#ffffff", true},
		{"#FFFFFF", true},
		{"#ff0000ff", true},
		{"#abc123", true},
		{"", false},
		{"fff", false},
		{"#gggggg", false},
		{"#ff", false},
	}
	for _, tt := range tests {
		r := ValidateHexColor(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateHexColor(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"40.7128,-74.0060", true},
		{"40.7128 -74.0060", true},
		{"0,0", true},
		{"90,180", true},
		{"-90,-180", true},
		{"91,0", false},
		{"0,181", false},
		{"", false},
		{"abc,def", false},
	}
	for _, tt := range tests {
		r := ValidateCoordinates(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateCoordinates(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"00000000-0000-1000-8000-000000000000", true},
		{"", false},
		{"550e8400-e29b-41d4-a716", false},
		{"550e8400-e29b-41d4-a716-44665544000g", false},
	}
	for _, tt := range tests {
		r := ValidateUUID(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateUUID(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"256.1.1.1", false},
		{"1.2.3", false},
		{"", false},
		{"::1", false},
	}
	for _, tt := range tests {
		r := ValidateIPv4(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateIPv4(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateIPv6(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		{"::", true},
		{"192.168.1.1", false},
		{"", false},
		{"gggg::1", false},
	}
	for _, tt := range tests {
		r := ValidateIPv6(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateIPv6(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"https://example.com", true},
		{"http://example.com/path", true},
		{"ftp://files.example.com", true},
		{"https://example.com:8080/path?q=1", true},
		{"", false},
		{"example.com", false},
		{"https://", false},
	}
	for _, tt := range tests {
		r := ValidateURL(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateURL(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateBase64(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"SGVsbG8gV29ybGQ=", true},
		{"SGVsbG8gV29ybGQ", true},
		{"-_-", true},
		{"", false},
		{"!!!notbase64!!!", false},
	}
	for _, tt := range tests {
		r := ValidateBase64(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateBase64(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateHex(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"deadbeef", true},
		{"0xdeadbeef", true},
		{"DEADBEEF", true},
		{"", false},
		{"xyz", false},
		{"abc", false},
	}
	for _, tt := range tests {
		r := ValidateHex(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateHex(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{`{"key": "value"}`, true},
		{`[1, 2, 3]`, true},
		{`"string"`, true},
		{`42`, true},
		{`null`, true},
		{``, false},
		{`{invalid}`, false},
	}
	for _, tt := range tests {
		r := ValidateJSON(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateJSON(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"4242424242424242", true},
		{"79927398713", true},
		{"79927398710", false},
		{"", false},
		{"abc", false},
	}
	for _, tt := range tests {
		r := ValidateLuhn(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateLuhn(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"hello-world", true},
		{"hello", true},
		{"hello-world-123", true},
		{"", false},
		{"Hello-World", false},
		{"-hello", false},
		{"hello-", false},
		{"hello--world", false},
	}
	for _, tt := range tests {
		r := ValidateSlug(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateSlug(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateBoolean(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"true", true},
		{"false", true},
		{"1", true},
		{"0", true},
		{"yes", true},
		{"no", true},
		{"on", true},
		{"off", true},
		{"", false},
		{"maybe", false},
	}
	for _, tt := range tests {
		r := ValidateBoolean(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateBoolean(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"john", true},
		{"john_doe", true},
		{"john-doe", true},
		{"user123", true},
		{"", false},
		{"ab", false},
		{"1john", false},
		{"_john", false},
		{"john__doe", false},
	}
	for _, tt := range tests {
		r := ValidateUsername(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateUsername(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestAutoDetect(t *testing.T) {
	tests := []struct {
		input    string
		valid    bool
	}{
		{"test@example.com", true},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"192.168.1.1", true},
		{"https://example.com", true},
		{"#ff0000", true},
		{"", false},
	}
	for _, tt := range tests {
		r := AutoDetect(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("AutoDetect(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateEIN(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12-3456789", true},
		{"123456789", true},
		{"", false},
		{"12345678", false},
		{"abc", false},
	}
	for _, tt := range tests {
		r := ValidateEIN(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateEIN(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateMIME(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"text/plain", true},
		{"application/json", true},
		{"image/png", true},
		{"text/html; charset=utf-8", false},
		{"", false},
		{"notamime", false},
	}
	for _, tt := range tests {
		r := ValidateMIME(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateMIME(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateASCII(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"Hello World", true},
		{"123!@#", true},
		{"", false},
		{"Hello\xc3\xa9", false},
	}
	for _, tt := range tests {
		r := ValidateASCII(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateASCII(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"42", true},
		{"3.14", true},
		{"-100", true},
		{"1e10", true},
		{"", false},
		{"abc", false},
	}
	for _, tt := range tests {
		r := ValidateNumeric(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateNumeric(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"Hello123", true},
		{"ABC", true},
		{"123", true},
		{"", false},
		{"hello world", false},
		{"hello!", false},
	}
	for _, tt := range tests {
		r := ValidateAlphanumeric(tt.input)
		if r.Valid != tt.valid {
			t.Errorf("ValidateAlphanumeric(%q) = %v, want %v, detail: %s", tt.input, r.Valid, tt.valid, r.Detail)
		}
	}
}
