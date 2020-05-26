package lurker

import "github.com/portmapping/lurker/proxy"

type localProxy struct {
	Config Proxy
	proxy.Proxy
}

// RegisterLocalProxy ...
func RegisterLocalProxy(l Lurker, cfg *Config) (err error) {
	for _, p := range cfg.Proxy {

		lp, err := proxy.New(p.Type, proxy.Auth{
			Name: p.Name,
			Pass: p.Pass,
		})
		if err != nil {
			return err
		}
		l.RegisterListener("", &localProxy{
			Config: p,
			Proxy:  lp,
		})
	}

	return nil
}

// Listen ...
func (p *localProxy) Listen(c chan<- Connector) (err error) {

	panic("implement me")

}

// Stop ...
func (p *localProxy) Stop() error {
	panic("implement me")
}

// IsReady ...
func (p *localProxy) IsReady() bool {
	panic("implement me")
}
