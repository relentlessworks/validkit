package config

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
)

// Config holds all configuration for the service.
type Config struct {
	Addr   string
	Secret string
	NoAuth bool
}

// Load reads configuration from defaults < env < flags.
func Load() *Config {
	c := &Config{
		Addr:   ":7700",
		Secret: "",
		NoAuth: false,
	}

	// Env
	if v := os.Getenv("VALIDKIT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("VALIDKIT_SECRET"); v != "" {
		c.Secret = v
	}
	if v := os.Getenv("VALIDKIT_NO_AUTH"); v == "1" || v == "true" {
		c.NoAuth = true
	}

	// Flags
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.StringVar(&c.Secret, "secret", c.Secret, "token signing secret (auto-generated if empty)")
	flag.BoolVar(&c.NoAuth, "no-auth", c.NoAuth, "disable authentication (dev mode)")
	flag.Parse()

	// Generate random secret if not provided
	if c.Secret == "" {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 32)
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		c.Secret = string(b)
		fmt.Fprintf(os.Stderr, "validkit: generated random secret (set VALIDKIT_SECRET to persist)\n")
	}

	return c
}
