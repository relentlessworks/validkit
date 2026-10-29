# validkit

> Agentic-first data validation service. Validate emails, credit cards, ISBNs, IBANs, BIC/SWIFT, VINs, SSNs, EINs, domains, MAC addresses, postal codes (28 countries), hex colors, coordinates, MIME types, usernames, JWT, JSON, XML, Base64, hex, ASCII, UUIDs, IPv4/IPv6, URLs, Luhn checksums, and more. Auto-detect validation type. Plain text API, agent-driven, single Go binary. No database needed — pure stateless computation.

## Quick Start

```bash
# Build
make build

# Run (dev mode, no auth)
./validkit -no-auth

# Validate an email
curl -X POST http://localhost:7700/validate/email -d '{"value":"test@example.com"}'
# valid=true type=email input=test@example.com detail=valid email address local=test domain=example.com

# Auto-detect type
curl -X POST http://localhost:7700/validate -d '{"value":"4242-4242-4242-4242"}'
# valid=true type=creditcard input=4242424242424242 detail=valid credit card number (Luhn check passed) type=visa

# Validate postal code with country
curl -X POST http://localhost:7700/validate/postal -d '{"value":"10001","country":"US"}'
# valid=true type=postal input=10001 detail=valid US ZIP code (5 digits or ZIP+4) country=US

# Get JSON output
curl -X POST http://localhost:7700/validate/uuid -H 'Accept: application/json' -d '{"value":"550e8400-e29b-41d4-a716-446655440000"}'

# List all validation types
curl http://localhost:7700/types

# Batch validation
curl -X POST http://localhost:7700/batch -d '{"items":[{"type":"email","value":"a@b.com"},{"type":"uuid","value":"550e8400-e29b-41d4-a716-446655440000"}]}'

# MCP endpoint
curl -X POST http://localhost:7700/mcp -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## Validation Types

| Type | Endpoint | Description |
|------|----------|-------------|
| email | /validate/email | Email address (RFC 5322) |
| creditcard | /validate/creditcard | Credit card (Luhn + type detection) |
| isbn | /validate/isbn | ISBN-10 or ISBN-13 |
| iban | /validate/iban | IBAN (mod-97 checksum) |
| bic | /validate/bic | BIC/SWIFT code |
| vin | /validate/vin | Vehicle Identification Number (17 chars, checksum) |
| ssn | /validate/ssn | US Social Security Number |
| ein | /validate/ein | US Employer Identification Number |
| domain | /validate/domain | Domain name |
| mac | /validate/mac | MAC address |
| postal | /validate/postal | Postal code (28 countries) |
| color | /validate/color | Hex color (#RGB, #RGBA, #RRGGBB, #RRGGBBAA) |
| coordinates | /validate/coordinates | Latitude/longitude coordinates |
| mime | /validate/mime | MIME type |
| username | /validate/username | Username/handle |
| jwt | /validate/jwt | JWT structure (header.payload.signature) |
| json | /validate/json | JSON validity |
| xml | /validate/xml | XML well-formedness |
| base64 | /validate/base64 | Base64 string (standard, url-safe, raw variants) |
| hex | /validate/hex | Hex string |
| ascii | /validate/ascii | ASCII string |
| uuid | /validate/uuid | UUID (version detection) |
| ipv4 | /validate/ipv4 | IPv4 address (class, private detection) |
| ipv6 | /validate/ipv6 | IPv6 address |
| url | /validate/url | URL |
| luhn | /validate/luhn | Luhn checksum |
| slug | /validate/slug | URL slug |
| boolean | /validate/boolean | Boolean value |
| numeric | /validate/numeric | Numeric string |
| alphanumeric | /validate/alphanumeric | Alphanumeric string |
| utf8 | /validate/utf8 | UTF-8 string |
| auto | /validate | Auto-detect type |

## Configuration

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| -addr | VALIDKIT_ADDR | :7700 | Listen address |
| -secret | VALIDKIT_SECRET | (auto) | Token signing secret |
| -no-auth | VALIDKIT_NO_AUTH | false | Disable auth (dev mode) |

## Build

```bash
make build    # CGO_ENABLED=0, single static binary
make test     # go test -race
make vet      # go vet
make run      # build + run with -no-auth
```

## Principles

- **Agent IS the interface** — No UI, no SDK. The API is the product.
- **Plain text by default** — One labeled, grepable line per result. JSON on demand.
- **Instructive errors** — Every 4xx includes a hint for self-correction.
- **Self-documenting** — `GET /help` returns the operating manual.
- **Single static binary** — Go, CGO_ENABLED=0, zero external dependencies.
- **No database** — Pure stateless computation.

## License

MIT
