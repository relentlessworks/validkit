package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/validkit/internal/model"
)

// MCPRequest represents a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{}  `json:"result,omitempty"`
	Error   *MCPCallError `json:"error,omitempty"`
}

// MCPCallError represents a JSON-RPC 2.0 error.
type MCPCallError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPTool represents a tool in the MCP protocol.
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema MCPInputSchema `json:"inputSchema"`
}

// MCPInputSchema represents the input schema for a tool.
type MCPInputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]MCPProperty `json:"properties"`
	Required   []string               `json:"required"`
}

// MCPProperty represents a property in the input schema.
type MCPProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use POST with JSON-RPC 2.0 body")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "failed to read body", "send a valid JSON-RPC 2.0 request")
		return
	}
	defer r.Body.Close()

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %s", err.Error()), "send a valid JSON-RPC 2.0 request")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "initialize":
		h.mcpInitialize(w, r, req)
	case "tools/list":
		h.mcpToolsList(w, r, req)
	case "tools/call":
		h.mcpToolsCall(w, r, req)
	default:
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPCallError{Code: -32601, Message: fmt.Sprintf("method '%s' not found", req.Method)},
		})
	}
}

func (h *Handler) mcpInitialize(w http.ResponseWriter, r *http.Request, req MCPRequest) {
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "validkit",
				"version": "0.1.0",
			},
		},
	})
}

func (h *Handler) mcpToolsList(w http.ResponseWriter, r *http.Request, req MCPRequest) {
	tools := mcpTools()
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	})
}

func (h *Handler) mcpToolsCall(w http.ResponseWriter, r *http.Request, req MCPRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPCallError{Code: -32602, Message: "invalid params"},
		})
		return
	}

	var args struct {
		Value   string `json:"value"`
		Country string `json:"country"`
	}
	if len(params.Arguments) > 0 {
		_ = json.Unmarshal(params.Arguments, &args)
	}

	var result model.ValidationResult
	switch strings.ToLower(params.Name) {
	case "validate_email":
		result = model.ValidateEmail(args.Value)
	case "validate_creditcard":
		result = model.ValidateCreditCard(args.Value)
	case "validate_isbn":
		result = model.ValidateISBN(args.Value)
	case "validate_iban":
		result = model.ValidateIBAN(args.Value)
	case "validate_bic":
		result = model.ValidateBIC(args.Value)
	case "validate_vin":
		result = model.ValidateVIN(args.Value)
	case "validate_ssn":
		result = model.ValidateSSN(args.Value)
	case "validate_ein":
		result = model.ValidateEIN(args.Value)
	case "validate_domain":
		result = model.ValidateDomain(args.Value)
	case "validate_mac":
		result = model.ValidateMAC(args.Value)
	case "validate_postal":
		result = model.ValidatePostal(args.Value, args.Country)
	case "validate_color":
		result = model.ValidateHexColor(args.Value)
	case "validate_coordinates":
		result = model.ValidateCoordinates(args.Value)
	case "validate_mime":
		result = model.ValidateMIME(args.Value)
	case "validate_username":
		result = model.ValidateUsername(args.Value)
	case "validate_jwt":
		result = model.ValidateJWT(args.Value)
	case "validate_json":
		result = model.ValidateJSON(args.Value)
	case "validate_xml":
		result = model.ValidateXML(args.Value)
	case "validate_base64":
		result = model.ValidateBase64(args.Value)
	case "validate_hex":
		result = model.ValidateHex(args.Value)
	case "validate_ascii":
		result = model.ValidateASCII(args.Value)
	case "validate_uuid":
		result = model.ValidateUUID(args.Value)
	case "validate_ipv4":
		result = model.ValidateIPv4(args.Value)
	case "validate_ipv6":
		result = model.ValidateIPv6(args.Value)
	case "validate_url":
		result = model.ValidateURL(args.Value)
	case "validate_luhn":
		result = model.ValidateLuhn(args.Value)
	case "validate_slug":
		result = model.ValidateSlug(args.Value)
	case "validate_boolean":
		result = model.ValidateBoolean(args.Value)
	case "validate_numeric":
		result = model.ValidateNumeric(args.Value)
	case "validate_alphanumeric":
		result = model.ValidateAlphanumeric(args.Value)
	case "validate_utf8":
		result = model.ValidateUTF8(args.Value)
	case "auto_detect":
		result = model.AutoDetect(args.Value)
	case "list_types":
		types := model.ListValidationTypes()
		var lines []string
		for _, t := range types {
			lines = append(lines, fmt.Sprintf("name=%s endpoint=%s description=%s", t.Name, t.Endpoint, t.Description))
		}
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": strings.Join(lines, "\n")},
				},
			},
		})
		return
	default:
		json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPCallError{Code: -32602, Message: fmt.Sprintf("unknown tool '%s'", params.Name)},
		})
		return
	}

	// Format result as text content
	var text string
	if result.Valid {
		text = fmt.Sprintf("valid=true type=%s input=%s detail=%s", result.Type, result.Input, result.Detail)
		for k, v := range result.Extra {
			text += fmt.Sprintf(" %s=%s", k, v)
		}
	} else {
		text = fmt.Sprintf("valid=false type=%s input=%s error=%s", result.Type, result.Input, result.Detail)
	}

	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": text},
			},
		},
	})
}

func mcpTools() []MCPTool {
	valueProp := MCPProperty{Type: "string", Description: "The value to validate"}
	countryProp := MCPProperty{Type: "string", Description: "Country code for postal validation (e.g. US, UK, CA)"}

	mkTool := func(name, desc string, extraProps map[string]MCPProperty, required []string) MCPTool {
		props := map[string]MCPProperty{"value": valueProp}
		for k, v := range extraProps {
			props[k] = v
		}
		return MCPTool{
			Name:        name,
			Description: desc,
			InputSchema: MCPInputSchema{
				Type:       "object",
				Properties: props,
				Required:   required,
			},
		}
	}

	tools := []MCPTool{
		mkTool("validate_email", "Validate an email address", nil, []string{"value"}),
		mkTool("validate_creditcard", "Validate a credit card number (Luhn + type detection)", nil, []string{"value"}),
		mkTool("validate_isbn", "Validate an ISBN-10 or ISBN-13", nil, []string{"value"}),
		mkTool("validate_iban", "Validate an IBAN", nil, []string{"value"}),
		mkTool("validate_bic", "Validate a BIC/SWIFT code", nil, []string{"value"}),
		mkTool("validate_vin", "Validate a VIN (Vehicle Identification Number)", nil, []string{"value"}),
		mkTool("validate_ssn", "Validate a US SSN", nil, []string{"value"}),
		mkTool("validate_ein", "Validate a US EIN", nil, []string{"value"}),
		mkTool("validate_domain", "Validate a domain name", nil, []string{"value"}),
		mkTool("validate_mac", "Validate a MAC address", nil, []string{"value"}),
		mkTool("validate_postal", "Validate a postal code", map[string]MCPProperty{"country": countryProp}, []string{"value"}),
		mkTool("validate_color", "Validate a hex color", nil, []string{"value"}),
		mkTool("validate_coordinates", "Validate lat/lng coordinates", nil, []string{"value"}),
		mkTool("validate_mime", "Validate a MIME type", nil, []string{"value"}),
		mkTool("validate_username", "Validate a username/handle", nil, []string{"value"}),
		mkTool("validate_jwt", "Validate JWT structure", nil, []string{"value"}),
		mkTool("validate_json", "Validate JSON", nil, []string{"value"}),
		mkTool("validate_xml", "Validate XML well-formedness", nil, []string{"value"}),
		mkTool("validate_base64", "Validate Base64 string", nil, []string{"value"}),
		mkTool("validate_hex", "Validate hex string", nil, []string{"value"}),
		mkTool("validate_ascii", "Validate ASCII string", nil, []string{"value"}),
		mkTool("validate_uuid", "Validate UUID", nil, []string{"value"}),
		mkTool("validate_ipv4", "Validate IPv4 address", nil, []string{"value"}),
		mkTool("validate_ipv6", "Validate IPv6 address", nil, []string{"value"}),
		mkTool("validate_url", "Validate URL", nil, []string{"value"}),
		mkTool("validate_luhn", "Validate Luhn checksum", nil, []string{"value"}),
		mkTool("validate_slug", "Validate slug", nil, []string{"value"}),
		mkTool("validate_boolean", "Validate boolean value", nil, []string{"value"}),
		mkTool("validate_numeric", "Validate numeric string", nil, []string{"value"}),
		mkTool("validate_alphanumeric", "Validate alphanumeric string", nil, []string{"value"}),
		mkTool("validate_utf8", "Validate UTF-8 string", nil, []string{"value"}),
		mkTool("auto_detect", "Auto-detect and validate input type", nil, []string{"value"}),
		mkTool("list_types", "List all validation types", nil, nil),
	}

	return tools
}
