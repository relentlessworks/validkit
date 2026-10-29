package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/validkit/internal/model"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	NoAuth bool
}

// NewHandler creates a new API handler.
func NewHandler(noAuth bool) *Handler {
	return &Handler{NoAuth: noAuth}
}

// RegisterRoutes registers all HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/types", h.listTypes)
	mux.HandleFunc("/postal/countries", h.listPostalCountries)
	mux.HandleFunc("/validate", h.validateAuto)
	mux.HandleFunc("/validate/", h.validateByPath)
	mux.HandleFunc("/batch", h.batch)
	mux.HandleFunc("/mcp", h.mcp)
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/", h.notFound)
}

// --- Response helpers ---

func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func writeResult(w http.ResponseWriter, r *http.Request, result model.ValidationResult) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if result.Valid {
		// Build plain text output: valid=true type=email input=... detail=... extra...
		parts := []string{
			fmt.Sprintf("valid=true"),
			fmt.Sprintf("type=%s", result.Type),
			fmt.Sprintf("input=%s", result.Input),
		}
		if result.Detail != "" {
			parts = append(parts, fmt.Sprintf("detail=%s", result.Detail))
		}
		for k, v := range result.Extra {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintln(w, strings.Join(parts, " "))
	} else {
		parts := []string{
			fmt.Sprintf("valid=false"),
			fmt.Sprintf("type=%s", result.Type),
			fmt.Sprintf("input=%s", result.Input),
		}
		if result.Detail != "" {
			parts = append(parts, fmt.Sprintf("error=%s", result.Detail))
		}
		fmt.Fprintln(w, strings.Join(parts, " "))
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, msg, hint string) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

func writeText(w http.ResponseWriter, r *http.Request, text string) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": text})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, text)
}

// --- Handlers ---

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeText(w, r, "ok")
}

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusNotFound, "unknown endpoint", "GET /help for the operating manual, GET /types for all validation types")
}

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	helpText := `validkit — agentic-first data validation service

Validate emails, credit cards, ISBNs, IBANs, BIC/SWIFT, VINs, SSNs, EINs,
domains, MAC addresses, postal codes (28 countries), hex colors, coordinates,
MIME types, usernames, JWT, JSON, XML, Base64, hex, ASCII, UUIDs, IPv4/IPv6,
URLs, Luhn checksums, and more. Auto-detect validation type from input.

== AUTH ==
This service supports optional auth. Use -no-auth flag or VALIDKIT_NO_AUTH=1
to disable auth for development. When auth is enabled, use:
  POST /auth/request  {email: "..."}  → sends OTP
  POST /auth/verify   {email: "...", code: "..."}  → returns bearer token
  Then send: Authorization: Bearer <token>

== ENDPOINTS ==

POST /validate
  Body: {"value": "test@example.com"}
  Auto-detects the type and validates. Returns valid=true/false with details.

POST /validate/<type>
  Body: {"value": "...", "country": "US"}  (country only for postal)
  Validates the input as the specified type.
  Types: email, creditcard, isbn, iban, bic, vin, ssn, ein, domain, mac,
         postal, color, coordinates, mime, username, jwt, json, xml,
         base64, hex, ascii, uuid, ipv4, ipv6, url, luhn, slug, boolean,
         numeric, alphanumeric, utf8, float, integer

GET /types
  Lists all supported validation types.

GET /postal/countries
  Lists supported postal code countries.

POST /batch
  Body: {"items": [{"type": "email", "value": "a@b.com"}, ...]}
  Validates multiple items at once.

POST /mcp
  MCP JSON-RPC 2.0 endpoint. Methods: initialize, tools/list, tools/call.

GET /help or /.well-known/agent.md
  This help text.

== RESPONSE FORMAT ==
Plain text (default): one labeled line per result
  valid=true type=email input=test@example.com detail=valid email address local=test domain=example.com
  valid=false type=email input=bad detail=not a valid email address

JSON (via Accept: application/json or ?format=json):
  {"valid": true, "type": "email", "input": "...", "detail": "...", "extra": {...}}

== ERRORS ==
4xx responses include a hint:
  error: missing 'value' field | hint: send JSON body like {"value": "test@example.com"}

== EXAMPLES ==
  curl -X POST http://localhost:7700/validate/email -d '{"value":"test@example.com"}'
  curl -X POST http://localhost:7700/validate -d '{"value":"4242-4242-4242-4242"}'
  curl -X POST 'http://localhost:7700/validate/postal' -d '{"value":"10001","country":"US"}'
  curl http://localhost:7700/types
  curl http://localhost:7700/validate/uuid?value=550e8400-e29b-41d4-a716-446655440000
`
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"help": helpText})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, helpText)
}

func (h *Handler) listTypes(w http.ResponseWriter, r *http.Request) {
	types := model.ListValidationTypes()
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, t := range types {
		fmt.Fprintf(w, "name=%s endpoint=%s description=%s\n", t.Name, t.Endpoint, t.Description)
	}
}

func (h *Handler) listPostalCountries(w http.ResponseWriter, r *http.Request) {
	countries := model.ListPostalCountries()
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"countries": countries})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "countries=%s\n", strings.Join(countries, ","))
}

func (h *Handler) validateAuto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use POST with JSON body like {\"value\": \"test@example.com\"}")
		return
	}

	value, country, err := parseValidateBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "send JSON body like {\"value\": \"test@example.com\"} or use query param ?value=...")
		return
	}

	_ = country // not used for auto-detect
	result := model.AutoDetect(value)
	writeResult(w, r, result)
}

func (h *Handler) validateByPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use POST with JSON body or GET with ?value= query param")
		return
	}

	// Extract type from path: /validate/<type>
	path := strings.TrimPrefix(r.URL.Path, "/validate/")
	if path == "" {
		writeError(w, r, http.StatusBadRequest, "missing validation type", "use /validate/<type> where type is one of: email, creditcard, isbn, iban, bic, vin, ssn, ein, domain, mac, postal, color, coordinates, mime, username, jwt, json, xml, base64, hex, ascii, uuid, ipv4, ipv6, url, luhn, slug, boolean, numeric, alphanumeric, utf8, float, integer")
		return
	}

	value, country, err := parseValidateBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "send JSON body like {\"value\": \"...\", \"country\": \"US\"} or use query param ?value=...")
		return
	}

	var result model.ValidationResult
	switch strings.ToLower(path) {
	case "email":
		result = model.ValidateEmail(value)
	case "creditcard", "credit_card", "cc":
		result = model.ValidateCreditCard(value)
	case "isbn":
		result = model.ValidateISBN(value)
	case "iban":
		result = model.ValidateIBAN(value)
	case "bic", "swift":
		result = model.ValidateBIC(value)
	case "vin":
		result = model.ValidateVIN(value)
	case "ssn":
		result = model.ValidateSSN(value)
	case "ein":
		result = model.ValidateEIN(value)
	case "domain":
		result = model.ValidateDomain(value)
	case "mac":
		result = model.ValidateMAC(value)
	case "postal", "zipcode", "zip":
		result = model.ValidatePostal(value, country)
	case "color", "hex_color":
		result = model.ValidateHexColor(value)
	case "coordinates", "coords", "latlng":
		result = model.ValidateCoordinates(value)
	case "mime", "mimetype":
		result = model.ValidateMIME(value)
	case "username", "handle":
		result = model.ValidateUsername(value)
	case "jwt":
		result = model.ValidateJWT(value)
	case "json":
		result = model.ValidateJSON(value)
	case "xml":
		result = model.ValidateXML(value)
	case "base64":
		result = model.ValidateBase64(value)
	case "hex":
		result = model.ValidateHex(value)
	case "ascii":
		result = model.ValidateASCII(value)
	case "uuid":
		result = model.ValidateUUID(value)
	case "ipv4":
		result = model.ValidateIPv4(value)
	case "ipv6":
		result = model.ValidateIPv6(value)
	case "url":
		result = model.ValidateURL(value)
	case "luhn":
		result = model.ValidateLuhn(value)
	case "slug":
		result = model.ValidateSlug(value)
	case "boolean":
		result = model.ValidateBoolean(value)
	case "numeric":
		result = model.ValidateNumeric(value)
	case "alphanumeric":
		result = model.ValidateAlphanumeric(value)
	case "utf8":
		result = model.ValidateUTF8(value)
	case "float":
		result = model.ValidateFloat(value, 0, 0)
	case "integer":
		result = model.ValidateInteger(value, 0, 0)
	default:
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("unknown validation type '%s'", path), "GET /types to see all supported validation types")
		return
	}

	writeResult(w, r, result)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use POST with JSON body like {\"items\": [{\"type\": \"email\", \"value\": \"a@b.com\"}]}")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "failed to read request body", "ensure you're sending a valid JSON body")
		return
	}
	defer r.Body.Close()

	var req struct {
		Items []model.BatchItem `json:"items"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err.Error()), "send JSON like {\"items\": [{\"type\": \"email\", \"value\": \"a@b.com\"}]}")
		return
	}

	if len(req.Items) == 0 {
		writeError(w, r, http.StatusBadRequest, "no items provided", "send JSON like {\"items\": [{\"type\": \"email\", \"value\": \"a@b.com\"}]}")
		return
	}

	if len(req.Items) > 100 {
		writeError(w, r, http.StatusBadRequest, "too many items (max 100)", "split your batch into smaller chunks")
		return
	}

	results := model.ValidateBatch(req.Items)
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, result := range results {
		if result.Valid {
			parts := []string{"valid=true", fmt.Sprintf("type=%s", result.Type), fmt.Sprintf("input=%s", result.Input)}
			if result.Detail != "" {
				parts = append(parts, fmt.Sprintf("detail=%s", result.Detail))
			}
			fmt.Fprintln(w, strings.Join(parts, " "))
		} else {
			parts := []string{"valid=false", fmt.Sprintf("type=%s", result.Type), fmt.Sprintf("input=%s", result.Input)}
			if result.Detail != "" {
				parts = append(parts, fmt.Sprintf("error=%s", result.Detail))
			}
			fmt.Fprintln(w, strings.Join(parts, " "))
		}
	}
}

// parseValidateBody extracts value and country from the request.
// Supports JSON body, query params, and plain text body.
func parseValidateBody(r *http.Request) (value, country string, err error) {
	// Try query params first (for GET requests)
	if v := r.URL.Query().Get("value"); v != "" {
		value = v
		country = r.URL.Query().Get("country")
		return value, country, nil
	}

	// Try JSON body
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		return "", "", fmt.Errorf("failed to read request body")
	}
	r.Body.Close()

	if len(body) == 0 {
		return "", "", fmt.Errorf("missing 'value' field")
	}

	// Try JSON first
	var jsonBody struct {
		Value   string `json:"value"`
		Country string `json:"country"`
	}
	if json.Unmarshal(body, &jsonBody) == nil && jsonBody.Value != "" {
		return jsonBody.Value, jsonBody.Country, nil
	}

	// Try plain text
	text := strings.TrimSpace(string(body))
	if text != "" {
		return text, "", nil
	}

	return "", "", fmt.Errorf("missing 'value' field")
}
