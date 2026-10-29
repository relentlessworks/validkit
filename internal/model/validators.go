package model

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValidationResult holds the outcome of a validation check.
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Type    string `json:"type"`
	Input   string `json:"input"`
	Detail  string `json:"detail,omitempty"`
	Extra   map[string]string `json:"extra,omitempty"`
}

// NewValidResult creates a valid result.
func NewValidResult(vtype, input, detail string) ValidationResult {
	return ValidationResult{Valid: true, Type: vtype, Input: input, Detail: detail}
}

// NewInvalidResult creates an invalid result.
func NewInvalidResult(vtype, input, detail string) ValidationResult {
	return ValidationResult{Valid: false, Type: vtype, Input: input, Detail: detail}
}

// --- Email ---

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

func ValidateEmail(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("email", input, "empty input")
	}
	if len(input) > 254 {
		return NewInvalidResult("email", input, "email exceeds 254 characters")
	}
	// Use Go's mail.ParseAddress for a robust check
	addr, err := mail.ParseAddress(input)
	if err != nil || !emailRegex.MatchString(input) {
		return NewInvalidResult("email", input, "not a valid email address")
	}
	parts := strings.Split(addr.Address, "@")
	domain := parts[1]
	extra := map[string]string{"local": parts[0], "domain": domain}
	return ValidationResult{Valid: true, Type: "email", Input: input, Detail: "valid email address", Extra: extra}
}

// --- Credit Card ---

func ValidateCreditCard(input string) ValidationResult {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "-", "")
	if input == "" {
		return NewInvalidResult("creditcard", input, "empty input")
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(input) {
		return NewInvalidResult("creditcard", input, "contains non-digit characters")
	}
	if len(input) < 13 || len(input) > 19 {
		return NewInvalidResult("creditcard", input, fmt.Sprintf("length %d is outside valid range (13-19)", len(input)))
	}
	if !luhnCheck(input) {
		return NewInvalidResult("creditcard", input, "fails Luhn checksum")
	}
	cardType := detectCardType(input)
	extra := map[string]string{"type": cardType}
	return ValidationResult{Valid: true, Type: "creditcard", Input: input, Detail: "valid credit card number (Luhn check passed)", Extra: extra}
}

func luhnCheck(num string) bool {
	sum := 0
	alt := false
	for i := len(num) - 1; i >= 0; i-- {
		d := int(num[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}

func detectCardType(num string) string {
	switch {
	case strings.HasPrefix(num, "4"):
		return "visa"
	case len(num) >= 2 && num[0:2] >= "51" && num[0:2] <= "55":
		return "mastercard"
	case len(num) >= 2 && (num[0:2] == "34" || num[0:2] == "37"):
		return "amex"
	case strings.HasPrefix(num, "6011") || strings.HasPrefix(num, "65") || (len(num) >= 6 && num[0:6] >= "622126" && num[0:6] <= "622925"):
		return "discover"
	case len(num) >= 4 && (num[0:4] == "6011" || num[0:4] == "6565"):
		return "discover"
	case strings.HasPrefix(num, "3528") || strings.HasPrefix(num, "3529") || strings.HasPrefix(num, "353") || strings.HasPrefix(num, "354") || strings.HasPrefix(num, "355") || strings.HasPrefix(num, "356") || strings.HasPrefix(num, "357") || strings.HasPrefix(num, "358"):
		return "jcb"
	case len(num) >= 3 && num[0:3] == "300" || len(num) >= 3 && num[0:3] == "301" || len(num) >= 3 && num[0:3] == "302" || len(num) >= 3 && num[0:3] == "303" || len(num) >= 3 && num[0:3] == "304" || len(num) >= 3 && num[0:3] == "305" || len(num) >= 4 && num[0:4] == "3095" || len(num) >= 4 && num[0:4] == "36" || len(num) >= 4 && num[0:4] == "38" || len(num) >= 4 && num[0:4] == "39":
		return "diners"
	case strings.HasPrefix(num, "62"):
		return "unionpay"
	default:
		return "unknown"
	}
}

// --- ISBN ---

func ValidateISBN(input string) ValidationResult {
	input = strings.ReplaceAll(input, "-", "")
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ToUpper(input)
	if input == "" {
		return NewInvalidResult("isbn", input, "empty input")
	}
	if len(input) == 10 {
		return validateISBN10(input)
	}
	if len(input) == 13 {
		return validateISBN13(input)
	}
	return NewInvalidResult("isbn", input, fmt.Sprintf("length %d is not 10 or 13", len(input)))
}

func validateISBN10(input string) ValidationResult {
	sum := 0
	for i := 0; i < 9; i++ {
		if input[i] < '0' || input[i] > '9' {
			return NewInvalidResult("isbn", input, fmt.Sprintf("invalid character '%c' at position %d", input[i], i+1))
		}
		sum += int(input[i]-'0') * (10 - i)
	}
	last := input[9]
	if last == 'X' {
		sum += 10
	} else if last >= '0' && last <= '9' {
		sum += int(last - '0')
	} else {
		return NewInvalidResult("isbn", input, fmt.Sprintf("invalid check digit '%c'", last))
	}
	if sum%11 != 0 {
		return NewInvalidResult("isbn", input, "fails ISBN-10 checksum")
	}
	extra := map[string]string{"format": "isbn-10"}
	return ValidationResult{Valid: true, Type: "isbn", Input: input, Detail: "valid ISBN-10", Extra: extra}
}

func validateISBN13(input string) ValidationResult {
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(input) {
		return NewInvalidResult("isbn", input, "contains non-digit characters")
	}
	sum := 0
	for i := 0; i < 12; i++ {
		d := int(input[i] - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	check := (10 - (sum % 10)) % 10
	if int(input[12]-'0') != check {
		return NewInvalidResult("isbn", input, "fails ISBN-13 checksum")
	}
	extra := map[string]string{"format": "isbn-13"}
	return ValidationResult{Valid: true, Type: "isbn", Input: input, Detail: "valid ISBN-13", Extra: extra}
}

// --- IBAN ---

func ValidateIBAN(input string) ValidationResult {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ToUpper(input)
	if input == "" {
		return NewInvalidResult("iban", input, "empty input")
	}
	if len(input) < 15 || len(input) > 34 {
		return NewInvalidResult("iban", input, fmt.Sprintf("length %d is outside valid range (15-34)", len(input)))
	}
	if !regexp.MustCompile(`^[A-Z]{2}[0-9A-Z]+$`).MatchString(input) {
		return NewInvalidResult("iban", input, "must start with 2-letter country code followed by alphanumeric characters")
	}
	// Rearrange: move first 4 chars to end
	rearranged := input[4:] + input[0:4]
	// Convert to integer string (A=10, B=11, ...)
	numStr := ""
	for _, ch := range rearranged {
		if ch >= '0' && ch <= '9' {
			numStr += string(ch)
		} else if ch >= 'A' && ch <= 'Z' {
			numStr += strconv.Itoa(int(ch - 'A' + 10))
		}
	}
	// Mod 97 check
	if mod97(numStr) != 1 {
		return NewInvalidResult("iban", input, "fails mod-97 checksum")
	}
	country := input[0:2]
	extra := map[string]string{"country": country}
	return ValidationResult{Valid: true, Type: "iban", Input: input, Detail: "valid IBAN", Extra: extra}
}

func mod97(numStr string) int {
	// Process in chunks to avoid overflow
	remainder := 0
	for i := 0; i < len(numStr); i++ {
		remainder = remainder*10 + int(numStr[i]-'0')
		if remainder >= 97 {
			remainder = remainder % 97
		}
	}
	return remainder
}

// --- BIC/SWIFT ---

func ValidateBIC(input string) ValidationResult {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return NewInvalidResult("bic", input, "empty input")
	}
	if !regexp.MustCompile(`^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$`).MatchString(input) {
		return NewInvalidResult("bic", input, "must be 8 or 11 characters: 6 letters + 2 alphanumeric (+ optional 3 alphanumeric)")
	}
	country := input[4:6]
	extra := map[string]string{"country": country, "format": "8"}
	if len(input) == 11 {
		extra["format"] = "11"
	}
	return ValidationResult{Valid: true, Type: "bic", Input: input, Detail: "valid BIC/SWIFT code", Extra: extra}
}

// --- VIN ---

func ValidateVIN(input string) ValidationResult {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return NewInvalidResult("vin", input, "empty input")
	}
	if len(input) != 17 {
		return NewInvalidResult("vin", input, fmt.Sprintf("length %d is not 17", len(input)))
	}
	// Check for invalid characters (I, O, Q are not allowed)
	for i, ch := range input {
		if ch == 'I' || ch == 'O' || ch == 'Q' {
			return NewInvalidResult("vin", input, fmt.Sprintf("invalid character '%c' at position %d (I, O, Q not allowed)", ch, i+1))
		}
	}
	// Transliteration values
	translit := map[rune]int{
		'A': 1, 'B': 2, 'C': 3, 'D': 4, 'E': 5, 'F': 6, 'G': 7, 'H': 8,
		'J': 1, 'K': 2, 'L': 3, 'M': 4, 'N': 5, 'P': 7, 'R': 9,
		'S': 2, 'T': 3, 'U': 4, 'V': 5, 'W': 6, 'X': 7, 'Y': 8, 'Z': 9,
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	}
	weights := []int{8, 7, 6, 5, 4, 3, 2, 10, 0, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i, ch := range input {
		val, ok := translit[ch]
		if !ok {
			return NewInvalidResult("vin", input, fmt.Sprintf("invalid character '%c' at position %d", ch, i+1))
		}
		sum += val * weights[i]
	}
	check := sum % 11
	if check == 10 {
		checkChar := 'X'
		if rune(input[8]) != checkChar {
			return NewInvalidResult("vin", input, fmt.Sprintf("check digit should be 'X' but got '%c'", input[8]))
		}
	} else {
		if int(input[8]-'0') != check {
			return NewInvalidResult("vin", input, fmt.Sprintf("check digit should be '%d' but got '%c'", check, input[8]))
		}
	}
	wmi := input[0:3]
	extra := map[string]string{"wmi": wmi}
	return ValidationResult{Valid: true, Type: "vin", Input: input, Detail: "valid VIN (17 chars, checksum passed)", Extra: extra}
}

// --- SSN (US) ---

func ValidateSSN(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("ssn", input, "empty input")
	}
	// Accept formats: XXX-XX-XXXX or XXXXXXXXX
	cleaned := strings.ReplaceAll(input, "-", "")
	if len(cleaned) != 9 {
		return NewInvalidResult("ssn", input, "must be 9 digits (XXX-XX-XXXX or XXXXXXXXX)")
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(cleaned) {
		return NewInvalidResult("ssn", input, "contains non-digit characters")
	}
	area, _ := strconv.Atoi(cleaned[0:3])
	group, _ := strconv.Atoi(cleaned[3:5])
	serial, _ := strconv.Atoi(cleaned[5:9])
	if area == 0 || area == 666 || area >= 900 {
		return NewInvalidResult("ssn", input, fmt.Sprintf("invalid area number %d (must be 001-899, not 666)", area))
	}
	if group == 0 {
		return NewInvalidResult("ssn", input, "invalid group number 0 (must be 01-99)")
	}
	if serial == 0 {
		return NewInvalidResult("ssn", input, "invalid serial number 0 (must be 0001-9999)")
	}
	// Check for obviously fake patterns
	if cleaned == "123456789" || cleaned == "000000000" {
		return NewInvalidResult("ssn", input, "obviously invalid SSN pattern")
	}
	extra := map[string]string{"area": cleaned[0:3], "group": cleaned[3:5], "serial": cleaned[5:9]}
	return ValidationResult{Valid: true, Type: "ssn", Input: input, Detail: "valid US SSN format", Extra: extra}
}

// --- EIN (US) ---

func ValidateEIN(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("ein", input, "empty input")
	}
	cleaned := strings.ReplaceAll(input, "-", "")
	if len(cleaned) != 9 {
		return NewInvalidResult("ein", input, "must be 9 digits (XX-XXXXXXX or XXXXXXXXX)")
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(cleaned) {
		return NewInvalidResult("ein", input, "contains non-digit characters")
	}
	prefix := cleaned[0:2]
	extra := map[string]string{"prefix": prefix}
	return ValidationResult{Valid: true, Type: "ein", Input: input, Detail: "valid EIN format", Extra: extra}
}

// --- Domain ---

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func ValidateDomain(input string) ValidationResult {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return NewInvalidResult("domain", input, "empty input")
	}
	if len(input) > 253 {
		return NewInvalidResult("domain", input, "domain exceeds 253 characters")
	}
	if !domainRegex.MatchString(input) {
		return NewInvalidResult("domain", input, "not a valid domain name")
	}
	parts := strings.Split(input, ".")
	tld := parts[len(parts)-1]
	extra := map[string]string{"tld": tld, "subdomains": strconv.Itoa(len(parts) - 1)}
	return ValidationResult{Valid: true, Type: "domain", Input: input, Detail: "valid domain name", Extra: extra}
}

// --- MAC Address ---

func ValidateMAC(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("mac", input, "empty input")
	}
	// Try parsing with net.ParseMAC
	_, err := net.ParseMAC(input)
	if err != nil {
		// net.ParseMAC doesn't accept bare 12-char hex, check manually
		if regexp.MustCompile(`^[0-9A-Fa-f]{12}$`).MatchString(input) {
			extra := map[string]string{"format": "bare"}
			return ValidationResult{Valid: true, Type: "mac", Input: input, Detail: "valid MAC address", Extra: extra}
		}
		return NewInvalidResult("mac", input, fmt.Sprintf("not a valid MAC address: %s", err.Error()))
	}
	// Determine format
	format := "unknown"
	if strings.Contains(input, ":") {
		format = "colon"
	} else if strings.Contains(input, "-") {
		format = "hyphen"
	} else if strings.Contains(input, ".") {
		format = "dot"
	} else {
		format = "bare"
	}
	extra := map[string]string{"format": format}
	return ValidationResult{Valid: true, Type: "mac", Input: input, Detail: "valid MAC address", Extra: extra}
}

// --- Postal Code ---

func ValidatePostal(input string, country string) ValidationResult {
	input = strings.TrimSpace(input)
	country = strings.ToUpper(strings.TrimSpace(country))
	if input == "" {
		return NewInvalidResult("postal", input, "empty input")
	}
	if country == "" {
		country = "US"
	}
	var regex *regexp.Regexp
	var desc string
	switch country {
	case "US":
		regex = regexp.MustCompile(`^\d{5}(-\d{4})?$`)
		desc = "US ZIP code (5 digits or ZIP+4)"
	case "UK", "GB":
		regex = regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`)
		desc = "UK postcode"
	case "CA":
		regex = regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`)
		desc = "Canadian postal code"
	case "DE":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "German postal code"
	case "FR":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "French postal code"
	case "AU":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Australian postal code"
	case "JP":
		regex = regexp.MustCompile(`^\d{3}-\d{4}$`)
		desc = "Japanese postal code"
	case "BR":
		regex = regexp.MustCompile(`^\d{5}-\d{3}$`)
		desc = "Brazilian CEP"
	case "IN":
		regex = regexp.MustCompile(`^\d{6}$`)
		desc = "Indian PIN code"
	case "NL":
		regex = regexp.MustCompile(`^\d{4}\s?[A-Z]{2}$`)
		desc = "Dutch postal code"
	case "ES":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "Spanish postal code"
	case "IT":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "Italian postal code"
	case "MX":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "Mexican postal code"
	case "PL":
		regex = regexp.MustCompile(`^\d{2}-\d{3}$`)
		desc = "Polish postal code"
	case "SE":
		regex = regexp.MustCompile(`^\d{3}\s?\d{2}$`)
		desc = "Swedish postal code"
	case "CH":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Swiss postal code"
	case "AT":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Austrian postal code"
	case "BE":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Belgian postal code"
	case "DK":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Danish postal code"
	case "FI":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "Finnish postal code"
	case "NO":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "Norwegian postal code"
	case "PT":
		regex = regexp.MustCompile(`^\d{4}-\d{3}$`)
		desc = "Portuguese postal code"
	case "CZ":
		regex = regexp.MustCompile(`^\d{3}\s?\d{2}$`)
		desc = "Czech postal code"
	case "RU":
		regex = regexp.MustCompile(`^\d{6}$`)
		desc = "Russian postal code"
	case "CN":
		regex = regexp.MustCompile(`^\d{6}$`)
		desc = "Chinese postal code"
	case "KR":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "South Korean postal code"
	case "ZA":
		regex = regexp.MustCompile(`^\d{4}$`)
		desc = "South African postal code"
	case "TR":
		regex = regexp.MustCompile(`^\d{5}$`)
		desc = "Turkish postal code"
	default:
		return NewInvalidResult("postal", input, fmt.Sprintf("unsupported country code '%s' — supported: US, UK, CA, DE, FR, AU, JP, BR, IN, NL, ES, IT, MX, PL, SE, CH, AT, BE, DK, FI, NO, PT, CZ, RU, CN, KR, ZA, TR", country))
	}
	inputUpper := strings.ToUpper(input)
	if !regex.MatchString(inputUpper) {
		return NewInvalidResult("postal", input, fmt.Sprintf("does not match %s format", desc))
	}
	extra := map[string]string{"country": country}
	return ValidationResult{Valid: true, Type: "postal", Input: input, Detail: fmt.Sprintf("valid %s", desc), Extra: extra}
}

// --- Hex Color ---

func ValidateHexColor(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("color", input, "empty input")
	}
	if !strings.HasPrefix(input, "#") {
		return NewInvalidResult("color", input, "must start with '#'")
	}
	hex := input[1:]
	if !regexp.MustCompile(`^[0-9A-Fa-f]+$`).MatchString(hex) {
		return NewInvalidResult("color", input, "contains non-hex characters")
	}
	format := "unknown"
	switch len(hex) {
	case 3:
		format = "#RGB"
	case 4:
		format = "#RGBA"
	case 6:
		format = "#RRGGBB"
	case 8:
		format = "#RRGGBBAA"
	default:
		return NewInvalidResult("color", input, fmt.Sprintf("invalid length %d after '#' (expected 3, 4, 6, or 8 hex digits)", len(hex)))
	}
	extra := map[string]string{"format": format}
	return ValidationResult{Valid: true, Type: "color", Input: input, Detail: "valid hex color", Extra: extra}
}

// --- Coordinates ---

func ValidateCoordinates(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("coordinates", input, "empty input")
	}
	// Accept "lat,lng" or "lat lng" format
	parts := regexp.MustCompile(`[,\s]+`).Split(input, -1)
	if len(parts) != 2 {
		return NewInvalidResult("coordinates", input, "must be 'lat,lng' or 'lat lng' format")
	}
	lat, err1 := strconv.ParseFloat(parts[0], 64)
	lng, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return NewInvalidResult("coordinates", input, "latitude and longitude must be numeric")
	}
	if lat < -90 || lat > 90 {
		return NewInvalidResult("coordinates", input, fmt.Sprintf("latitude %g is outside valid range (-90 to 90)", lat))
	}
	if lng < -180 || lng > 180 {
		return NewInvalidResult("coordinates", input, fmt.Sprintf("longitude %g is outside valid range (-180 to 180)", lng))
	}
	extra := map[string]string{
		"latitude":  fmt.Sprintf("%g", lat),
		"longitude": fmt.Sprintf("%g", lng),
	}
	return ValidationResult{Valid: true, Type: "coordinates", Input: input, Detail: "valid geographic coordinates", Extra: extra}
}

// --- MIME Type ---

func ValidateMIME(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("mime", input, "empty input")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9!#$&^_.+-]+/[a-zA-Z0-9!#$&^_.+-]+$`).MatchString(input) {
		return NewInvalidResult("mime", input, "not a valid MIME type (expected type/subtype format)")
	}
	parts := strings.Split(input, "/")
	extra := map[string]string{"type": parts[0], "subtype": parts[1]}
	return ValidationResult{Valid: true, Type: "mime", Input: input, Detail: "valid MIME type format", Extra: extra}
}

// --- Username ---

func ValidateUsername(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("username", input, "empty input")
	}
	if len(input) < 3 {
		return NewInvalidResult("username", input, "too short (minimum 3 characters)")
	}
	if len(input) > 30 {
		return NewInvalidResult("username", input, "too long (maximum 30 characters)")
	}
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(input) {
		return NewInvalidResult("username", input, "must start with a letter and contain only letters, digits, underscores, and hyphens")
	}
	if strings.Contains(input, "__") || strings.HasPrefix(input, "_") || strings.HasSuffix(input, "_") {
		return NewInvalidResult("username", input, "cannot start/end with underscore or contain consecutive underscores")
	}
	if strings.Contains(input, "--") || strings.HasPrefix(input, "-") || strings.HasSuffix(input, "-") {
		return NewInvalidResult("username", input, "cannot start/end with hyphen or contain consecutive hyphens")
	}
	return ValidationResult{Valid: true, Type: "username", Input: input, Detail: "valid username"}
}

// --- JWT ---

func ValidateJWT(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("jwt", input, "empty input")
	}
	parts := strings.Split(input, ".")
	if len(parts) != 3 {
		return NewInvalidResult("jwt", input, fmt.Sprintf("expected 3 parts separated by '.', got %d", len(parts)))
	}
	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return NewInvalidResult("jwt", input, fmt.Sprintf("invalid header encoding: %s", err.Error()))
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return NewInvalidResult("jwt", input, fmt.Sprintf("invalid header JSON: %s", err.Error()))
	}
	alg, _ := header["alg"].(string)
	if alg == "" {
		return NewInvalidResult("jwt", input, "header missing 'alg' field")
	}
	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return NewInvalidResult("jwt", input, fmt.Sprintf("invalid payload encoding: %s", err.Error()))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return NewInvalidResult("jwt", input, fmt.Sprintf("invalid payload JSON: %s", err.Error()))
	}
	// Check signature is non-empty base64url
	if len(parts[2]) == 0 {
		return NewInvalidResult("jwt", input, "signature is empty")
	}
	extra := map[string]string{"alg": alg}
	if typ, ok := header["typ"].(string); ok {
		extra["typ"] = typ
	}
	if iss, ok := payload["iss"].(string); ok {
		extra["issuer"] = iss
	}
	if sub, ok := payload["sub"].(string); ok {
		extra["subject"] = sub
	}
	return ValidationResult{Valid: true, Type: "jwt", Input: input, Detail: "valid JWT structure (header.payload.signature)", Extra: extra}
}

// --- JSON ---

func ValidateJSON(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("json", input, "empty input")
	}
	var v interface{}
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return NewInvalidResult("json", input, fmt.Sprintf("invalid JSON: %s", err.Error()))
	}
	return ValidationResult{Valid: true, Type: "json", Input: input, Detail: "valid JSON"}
}

// --- XML ---

func ValidateXML(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("xml", input, "empty input")
	}
	// Basic XML well-formedness check using encoding/xml
	// We use a simple parser approach
	if !strings.HasPrefix(input, "<") {
		return NewInvalidResult("xml", input, "must start with '<'")
	}
	// Use a regex-based approach for basic well-formedness
	// A proper XML parser would be better but we keep it dependency-free
	xmlRegex := regexp.MustCompile(`^<\?xml[^>]*\?>\s*<([a-zA-Z_][\w.-]*)([\s\S]*?)>([\s\S]*)</\1>\s*$|^<([a-zA-Z_][\w.-]*)([\s\S]*?)>([\s\S]*)</\4>\s*$`)
	if !xmlRegex.MatchString(input) {
		// Try a simpler check - just needs opening and closing tags
		if !strings.HasSuffix(strings.TrimSpace(input), ">") {
			return NewInvalidResult("xml", input, "must end with '>'")
		}
		// Check for balanced tags using a simple stack
		if !checkXMLWellFormed(input) {
			return NewInvalidResult("xml", input, "not well-formed XML (unbalanced tags)")
		}
	}
	return ValidationResult{Valid: true, Type: "xml", Input: input, Detail: "well-formed XML"}
}

func checkXMLWellFormed(input string) bool {
	tagRegex := regexp.MustCompile(`<(/?)([a-zA-Z_][\w.-]*)[^>]*?(/?)>`)
	stack := []string{}
	matches := tagRegex.FindAllStringSubmatch(input, -1)
	for _, m := range matches {
		isClosing := m[1] == "/"
		tagName := m[2]
		isSelfClosing := m[3] == "/"
		if isSelfClosing {
			continue
		}
		if isClosing {
			if len(stack) == 0 || stack[len(stack)-1] != tagName {
				return false
			}
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, tagName)
		}
	}
	return len(stack) == 0
}

// --- Base64 ---

func ValidateBase64(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("base64", input, "empty input")
	}
	// Try standard base64
	_, errStd := base64.StdEncoding.DecodeString(input)
	if errStd == nil {
		return ValidationResult{Valid: true, Type: "base64", Input: input, Detail: "valid standard Base64", Extra: map[string]string{"variant": "standard"}}
	}
	// Try URL-safe base64
	_, errURL := base64.URLEncoding.DecodeString(input)
	if errURL == nil {
		return ValidationResult{Valid: true, Type: "base64", Input: input, Detail: "valid URL-safe Base64", Extra: map[string]string{"variant": "url-safe"}}
	}
	// Try raw standard (no padding)
	_, errRawStd := base64.RawStdEncoding.DecodeString(input)
	if errRawStd == nil {
		return ValidationResult{Valid: true, Type: "base64", Input: input, Detail: "valid raw standard Base64 (no padding)", Extra: map[string]string{"variant": "raw-standard"}}
	}
	// Try raw URL-safe (no padding)
	_, errRawURL := base64.RawURLEncoding.DecodeString(input)
	if errRawURL == nil {
		return ValidationResult{Valid: true, Type: "base64", Input: input, Detail: "valid raw URL-safe Base64 (no padding)", Extra: map[string]string{"variant": "raw-url-safe"}}
	}
	return NewInvalidResult("base64", input, "not valid Base64 in any variant (standard, url-safe, raw-standard, raw-url-safe)")
}

// --- Hex String ---

func ValidateHex(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("hex", input, "empty input")
	}
	// Remove optional 0x prefix
	if strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "0X") {
		input = input[2:]
	}
	if len(input)%2 != 0 {
		return NewInvalidResult("hex", input, "odd number of hex digits")
	}
	if !regexp.MustCompile(`^[0-9A-Fa-f]+$`).MatchString(input) {
		return NewInvalidResult("hex", input, "contains non-hex characters")
	}
	_, err := hex.DecodeString(input)
	if err != nil {
		return NewInvalidResult("hex", input, fmt.Sprintf("invalid hex: %s", err.Error()))
	}
	return ValidationResult{Valid: true, Type: "hex", Input: input, Detail: fmt.Sprintf("valid hex string (%d bytes)", len(input)/2)}
}

// --- ASCII ---

func ValidateASCII(input string) ValidationResult {
	if input == "" {
		return NewInvalidResult("ascii", input, "empty input")
	}
	for i, ch := range input {
		if ch > 127 {
			return NewInvalidResult("ascii", input, fmt.Sprintf("non-ASCII character at position %d (code point U+%04X)", i+1, ch))
		}
	}
	return ValidationResult{Valid: true, Type: "ascii", Input: input, Detail: "valid ASCII string"}
}

// --- UUID ---

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ValidateUUID(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("uuid", input, "empty input")
	}
	if !uuidRegex.MatchString(input) {
		return NewInvalidResult("uuid", input, "must match format XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX (8-4-4-4-12 hex digits)")
	}
	// Determine version
	versionChar := input[14]
	version := ""
	switch versionChar {
	case '1':
		version = "v1 (time-based)"
	case '2':
		version = "v2 (DCE Security)"
	case '3':
		version = "v3 (name-based MD5)"
	case '4':
		version = "v4 (random)"
	case '5':
		version = "v5 (name-based SHA1)"
	case '6':
		version = "v6 (time-ordered)"
	case '7':
		version = "v7 (time-ordered random)"
	case '8':
		version = "v8 (custom)"
	default:
		return NewInvalidResult("uuid", input, fmt.Sprintf("invalid version digit '%c'", versionChar))
	}
	extra := map[string]string{"version": version}
	// Check variant (must be 8, 9, a, or b in position 19)
	variantChar := strings.ToLower(string(input[19]))
	if variantChar != "8" && variantChar != "9" && variantChar != "a" && variantChar != "b" {
		return NewInvalidResult("uuid", input, fmt.Sprintf("invalid variant character '%c' at position 20", input[19]))
	}
	return ValidationResult{Valid: true, Type: "uuid", Input: input, Detail: "valid UUID", Extra: extra}
}

// --- IPv4 ---

func ValidateIPv4(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("ipv4", input, "empty input")
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return NewInvalidResult("ipv4", input, "not a valid IP address")
	}
	if !strings.Contains(input, ".") {
		return NewInvalidResult("ipv4", input, "not an IPv4 address (no dots)")
	}
	// Verify it's actually IPv4
	v4 := ip.To4()
	if v4 == nil {
		return NewInvalidResult("ipv4", input, "not an IPv4 address")
	}
	extra := map[string]string{
		"class":    getIPv4Class(v4),
		"private":  fmt.Sprintf("%t", isPrivateIPv4(v4)),
	}
	return ValidationResult{Valid: true, Type: "ipv4", Input: input, Detail: "valid IPv4 address", Extra: extra}
}

func getIPv4Class(ip net.IP) string {
	first := ip[0]
	switch {
	case first >= 1 && first <= 126:
		return "A"
	case first == 127:
		return "loopback"
	case first >= 128 && first <= 191:
		return "B"
	case first >= 192 && first <= 223:
		return "C"
	case first >= 224 && first <= 239:
		return "D (multicast)"
	case first >= 240:
		return "E (reserved)"
	default:
		return "unknown"
	}
}

func isPrivateIPv4(ip net.IP) bool {
	if ip[0] == 10 {
		return true
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return true
	}
	if ip[0] == 192 && ip[1] == 168 {
		return true
	}
	return false
}

// --- IPv6 ---

func ValidateIPv6(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("ipv6", input, "empty input")
	}
	ip := net.ParseIP(input)
	if ip == nil {
		return NewInvalidResult("ipv6", input, "not a valid IP address")
	}
	if !strings.Contains(input, ":") {
		return NewInvalidResult("ipv6", input, "not an IPv6 address (no colons)")
	}
	v4 := ip.To4()
	if v4 != nil {
		return NewInvalidResult("ipv6", input, "this is an IPv4 address, not IPv6")
	}
	extra := map[string]string{
		"loopback": fmt.Sprintf("%t", ip.IsLoopback()),
		"link_local": fmt.Sprintf("%t", ip.IsLinkLocalUnicast()),
	}
	return ValidationResult{Valid: true, Type: "ipv6", Input: input, Detail: "valid IPv6 address", Extra: extra}
}

// --- URL ---

func ValidateURL(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("url", input, "empty input")
	}
	u, err := url.Parse(input)
	if err != nil {
		return NewInvalidResult("url", input, fmt.Sprintf("invalid URL: %s", err.Error()))
	}
	if u.Scheme == "" {
		return NewInvalidResult("url", input, "missing scheme (e.g. http://, https://)")
	}
	if u.Host == "" {
		return NewInvalidResult("url", input, "missing host")
	}
	extra := map[string]string{
		"scheme": u.Scheme,
		"host":   u.Host,
	}
	if u.Path != "" {
		extra["path"] = u.Path
	}
	return ValidationResult{Valid: true, Type: "url", Input: input, Detail: "valid URL", Extra: extra}
}

// --- Luhn ---

func ValidateLuhn(input string) ValidationResult {
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "-", "")
	if input == "" {
		return NewInvalidResult("luhn", input, "empty input")
	}
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(input) {
		return NewInvalidResult("luhn", input, "contains non-digit characters")
	}
	if luhnCheck(input) {
		return ValidationResult{Valid: true, Type: "luhn", Input: input, Detail: "passes Luhn checksum"}
	}
	return NewInvalidResult("luhn", input, "fails Luhn checksum")
}

// --- Auto-detect ---

func AutoDetect(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("unknown", input, "empty input")
	}

	// Try each validator in order of specificity
	// Credit card before JSON/hex/base64/luhn since a 16-digit number is valid JSON
	candidates := []struct {
		name string
		fn   func(string) ValidationResult
	}{
		{"uuid", ValidateUUID},
		{"ipv4", ValidateIPv4},
		{"ipv6", ValidateIPv6},
		{"mac", ValidateMAC},
		{"url", ValidateURL},
		{"email", ValidateEmail},
		{"hex_color", ValidateHexColor},
		{"jwt", ValidateJWT},
		{"creditcard", ValidateCreditCard},
		{"isbn", ValidateISBN},
		{"iban", ValidateIBAN},
		{"bic", ValidateBIC},
		{"vin", ValidateVIN},
		{"ssn", ValidateSSN},
		{"ein", ValidateEIN},
		{"domain", ValidateDomain},
		{"json", ValidateJSON},
		{"base64", ValidateBase64},
		{"hex", ValidateHex},
		{"mime", ValidateMIME},
		{"username", ValidateUsername},
		{"luhn", ValidateLuhn},
	}

	var matches []string
	var firstValid *ValidationResult
	for _, c := range candidates {
		result := c.fn(input)
		if result.Valid {
			matches = append(matches, c.name)
			if firstValid == nil {
				r := result
				firstValid = &r
			}
		}
	}

	if len(matches) == 0 {
		return NewInvalidResult("unknown", input, "could not auto-detect a valid type for this input")
	}

	if firstValid != nil {
		firstValid.Detail = fmt.Sprintf("auto-detected as %s — %s", matches[0], firstValid.Detail)
		firstValid.Extra = map[string]string{"detected_types": strings.Join(matches, ", ")}
		return *firstValid
	}

	return NewInvalidResult("unknown", input, "auto-detection failed")
}

// --- Batch ---

// BatchItem represents a single item in a batch validation request.
type BatchItem struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Country string `json:"country,omitempty"`
}

func ValidateBatch(items []BatchItem) []ValidationResult {
	results := make([]ValidationResult, len(items))
	for i, item := range items {
		results[i] = validateByType(item.Type, item.Value, item.Country)
	}
	return results
}

func validateByType(vtype, value, country string) ValidationResult {
	vtype = strings.ToLower(strings.TrimSpace(vtype))
	switch vtype {
	case "email":
		return ValidateEmail(value)
	case "creditcard", "credit_card", "cc":
		return ValidateCreditCard(value)
	case "isbn":
		return ValidateISBN(value)
	case "iban":
		return ValidateIBAN(value)
	case "bic", "swift":
		return ValidateBIC(value)
	case "vin":
		return ValidateVIN(value)
	case "ssn":
		return ValidateSSN(value)
	case "ein":
		return ValidateEIN(value)
	case "domain":
		return ValidateDomain(value)
	case "mac":
		return ValidateMAC(value)
	case "postal", "zipcode", "zip":
		return ValidatePostal(value, country)
	case "color", "hex_color":
		return ValidateHexColor(value)
	case "coordinates", "coords", "latlng":
		return ValidateCoordinates(value)
	case "mime", "mimetype":
		return ValidateMIME(value)
	case "username", "handle":
		return ValidateUsername(value)
	case "jwt":
		return ValidateJWT(value)
	case "json":
		return ValidateJSON(value)
	case "xml":
		return ValidateXML(value)
	case "base64":
		return ValidateBase64(value)
	case "hex":
		return ValidateHex(value)
	case "ascii":
		return ValidateASCII(value)
	case "uuid":
		return ValidateUUID(value)
	case "ipv4":
		return ValidateIPv4(value)
	case "ipv6":
		return ValidateIPv6(value)
	case "url":
		return ValidateURL(value)
	case "luhn":
		return ValidateLuhn(value)
	case "auto", "detect", "":
		return AutoDetect(value)
	default:
		return NewInvalidResult(vtype, value, fmt.Sprintf("unknown validation type '%s' — use GET /types to see supported types", vtype))
	}
}

// --- Validation Types ---

type ValidationType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
}

func ListValidationTypes() []ValidationType {
	return []ValidationType{
		{"email", "Validate email address (RFC 5322)", "/validate/email"},
		{"creditcard", "Validate credit card (Luhn + type detection)", "/validate/creditcard"},
		{"isbn", "Validate ISBN-10 or ISBN-13", "/validate/isbn"},
		{"iban", "Validate IBAN (mod-97)", "/validate/iban"},
		{"bic", "Validate BIC/SWIFT code", "/validate/bic"},
		{"vin", "Validate VIN (17 chars, checksum)", "/validate/vin"},
		{"ssn", "Validate US SSN", "/validate/ssn"},
		{"ein", "Validate US EIN", "/validate/ein"},
		{"domain", "Validate domain name", "/validate/domain"},
		{"mac", "Validate MAC address", "/validate/mac"},
		{"postal", "Validate postal code (28 countries)", "/validate/postal"},
		{"color", "Validate hex color (#RGB, #RGBA, #RRGGBB, #RRGGBBAA)", "/validate/color"},
		{"coordinates", "Validate lat/lng coordinates", "/validate/coordinates"},
		{"mime", "Validate MIME type", "/validate/mime"},
		{"username", "Validate username/handle", "/validate/username"},
		{"jwt", "Validate JWT structure", "/validate/jwt"},
		{"json", "Validate JSON", "/validate/json"},
		{"xml", "Validate XML well-formedness", "/validate/xml"},
		{"base64", "Validate Base64 string", "/validate/base64"},
		{"hex", "Validate hex string", "/validate/hex"},
		{"ascii", "Validate ASCII string", "/validate/ascii"},
		{"uuid", "Validate UUID", "/validate/uuid"},
		{"ipv4", "Validate IPv4 address", "/validate/ipv4"},
		{"ipv6", "Validate IPv6 address", "/validate/ipv6"},
		{"url", "Validate URL", "/validate/url"},
		{"luhn", "Validate Luhn checksum", "/validate/luhn"},
		{"auto", "Auto-detect and validate", "/validate"},
	}
}

// --- Supported Postal Countries ---

func ListPostalCountries() []string {
	return []string{
		"US", "UK", "CA", "DE", "FR", "AU", "JP", "BR", "IN", "NL",
		"ES", "IT", "MX", "PL", "SE", "CH", "AT", "BE", "DK", "FI",
		"NO", "PT", "CZ", "RU", "CN", "KR", "ZA", "TR",
	}
}

// --- Helper: generate random token (for potential future use) ---

func GenerateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// --- Helper: check if string is valid UTF-8 ---

func ValidateUTF8(input string) ValidationResult {
	if input == "" {
		return NewInvalidResult("utf8", input, "empty input")
	}
	if !utf8.ValidString(input) {
		return NewInvalidResult("utf8", input, "contains invalid UTF-8 sequences")
	}
	runeCount := utf8.RuneCountInString(input)
	extra := map[string]string{
		"bytes":     strconv.Itoa(len(input)),
		"runes":     strconv.Itoa(runeCount),
	}
	return ValidationResult{Valid: true, Type: "utf8", Input: input, Detail: "valid UTF-8 string", Extra: extra}
}

// --- Helper: check if string is numeric ---

func ValidateNumeric(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("numeric", input, "empty input")
	}
	if _, err := strconv.ParseFloat(input, 64); err != nil {
		return NewInvalidResult("numeric", input, "not a valid number")
	}
	return ValidationResult{Valid: true, Type: "numeric", Input: input, Detail: "valid numeric string"}
}

// --- Helper: check if string is alphanumeric ---

func ValidateAlphanumeric(input string) ValidationResult {
	if input == "" {
		return NewInvalidResult("alphanumeric", input, "empty input")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(input) {
		return NewInvalidResult("alphanumeric", input, "contains non-alphanumeric characters")
	}
	return ValidationResult{Valid: true, Type: "alphanumeric", Input: input, Detail: "valid alphanumeric string"}
}

// --- Helper: check if string is a valid slug ---

func ValidateSlug(input string) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("slug", input, "empty input")
	}
	if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(input) {
		return NewInvalidResult("slug", input, "must be lowercase alphanumeric with hyphens (no leading/trailing/consecutive hyphens)")
	}
	return ValidationResult{Valid: true, Type: "slug", Input: input, Detail: "valid slug"}
}

// --- Helper: check if value is a valid boolean ---

func ValidateBoolean(input string) ValidationResult {
	input = strings.ToLower(strings.TrimSpace(input))
	switch input {
	case "true", "false", "1", "0", "yes", "no", "on", "off":
		extra := map[string]string{"normalized": normalizeBool(input)}
		return ValidationResult{Valid: true, Type: "boolean", Input: input, Detail: "valid boolean value", Extra: extra}
	default:
		return NewInvalidResult("boolean", input, "not a valid boolean (accepted: true/false, 1/0, yes/no, on/off)")
	}
}

func normalizeBool(input string) string {
	switch input {
	case "true", "1", "yes", "on":
		return "true"
	case "false", "0", "no", "off":
		return "false"
	default:
		return input
	}
}

// --- Helper: check if value is a valid float within range ---

func ValidateFloat(input string, min, max float64) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("float", input, "empty input")
	}
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return NewInvalidResult("float", input, "not a valid float")
	}
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return NewInvalidResult("float", input, "value is NaN or Infinity")
	}
	if min != 0 || max != 0 {
		if val < min {
			return NewInvalidResult("float", input, fmt.Sprintf("value %g is below minimum %g", val, min))
		}
		if val > max {
			return NewInvalidResult("float", input, fmt.Sprintf("value %g is above maximum %g", val, max))
		}
	}
	extra := map[string]string{"value": fmt.Sprintf("%g", val)}
	return ValidationResult{Valid: true, Type: "float", Input: input, Detail: "valid float", Extra: extra}
}

// --- Helper: check if value is a valid integer within range ---

func ValidateInteger(input string, min, max int64) ValidationResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return NewInvalidResult("integer", input, "empty input")
	}
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return NewInvalidResult("integer", input, "not a valid integer")
	}
	if min != 0 || max != 0 {
		if val < min {
			return NewInvalidResult("integer", input, fmt.Sprintf("value %d is below minimum %d", val, min))
		}
		if val > max {
			return NewInvalidResult("integer", input, fmt.Sprintf("value %d is above maximum %d", val, max))
		}
	}
	extra := map[string]string{"value": strconv.FormatInt(val, 10)}
	return ValidationResult{Valid: true, Type: "integer", Input: input, Detail: "valid integer", Extra: extra}
}
