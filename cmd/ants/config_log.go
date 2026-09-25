package main

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
)

// safePrintConfig logs the given config as JSON, redacting any password-like
// fields (and credentials embedded in URL values) so secrets don't leak.
func safePrintConfig(tag string, cfg any) {
	dat, err := json.Marshal(cfg)
	if err != nil {
		slog.Warn("Failed marshalling config", "tag", tag, "err", err.Error())
		return
	}

	// Decode into a generic structure so we can redact secrets wherever they
	// are nested before printing the config.
	var generic any
	if err := json.Unmarshal(dat, &generic); err != nil {
		slog.Warn("Failed unmarshalling config", "tag", tag, "err", err.Error())
		return
	}
	redactSecrets(generic)

	dat, err = json.Marshal(generic)
	if err != nil {
		slog.Warn("Failed marshalling config", "tag", tag, "err", err.Error())
		return
	}

	slog.Info("config", "tag", tag, "config", string(dat))
}

// redactSecrets recursively replaces the value of any map key that looks like a
// password or secret with a placeholder, and masks credentials in URL values.
func redactSecrets(v any) {
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	for key, val := range m {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "pass") || strings.Contains(lower, "secret") {
			m[key] = "*****"
			continue
		}
		if s, ok := val.(string); ok {
			m[key] = redactURLCredentials(s)
			continue
		}
		redactSecrets(val)
	}
}

// redactURLCredentials masks the password in a URL's user:password so it doesn't
// leak into logs. Non-URL or passwordless strings are returned unchanged.
func redactURLCredentials(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	if _, hasPwd := u.User.Password(); !hasPwd {
		return raw
	}
	u.User = url.UserPassword(u.User.Username(), "....")
	return u.String()
}
