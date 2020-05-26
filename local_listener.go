package lurker

import (
	"errors"
)

type localSocks5Listener struct {
}

// Server ...
func Local(cfg *Config) (Listener, error) {
	if !cfg.UseProxy {
		return nil, errors.New("not supported")
	}
	if cfg.Proxy.Name != "" && cfg.Proxy.Pass != "" {
	}

	return nil
}
