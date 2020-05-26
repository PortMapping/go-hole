package lurker

import (
	"errors"
)

type baseLocalListener struct {
}
type localSocks5Listener struct {
	baseLocalListener
}

// RegisterLocalProxy ...
func RegisterLocalProxy(l Lurker, cfg *Config) (err error) {
	if !cfg.UseProxy {
		return errors.New("not supported")
	}
	for _, proxy := range cfg.Proxy {
		if proxy.Pass != "" && proxy.Name != "" {

		}
	}
	return nil
}
