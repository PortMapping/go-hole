package lurker

import "net"

// Server ...
func Local(cfg *Config, l *config.LocalServer, config *config.CommonConfig) error {
	if l.Type != "secret" {
		go handleUdpMonitor(config, l)
	}
	task := &file.Tunnel{
		Port:     l.Port,
		ServerIp: "0.0.0.0",
		Status:   true,
		Client: &file.Client{
			Cnf: &file.Config{
				U:        "",
				P:        "",
				Compress: config.Client.Cnf.Compress,
			},
			Status:    true,
			RateLimit: 0,
			Flow:      &file.Flow{},
		},
		Flow:   &file.Flow{},
		Target: &file.Target{},
	}
	switch l.Type {
	case "p2ps":
		logs.Info("successful start-up of local socks5 monitoring, port", l.Port)
		return proxy.NewSock5ModeServer(p2pNetBridge, task).Start()
	case "p2pt":
		logs.Info("successful start-up of local tcp trans monitoring, port", l.Port)
		return proxy.NewTunnelModeServer(proxy.HandleTrans, p2pNetBridge, task).Start()
	case "p2p", "secret":
		listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), l.Port, ""})
		if err != nil {
			logs.Error("local listener startup failed port %d, error %s", l.Port, err.Error())
			return err
		}
		LocalServer = append(LocalServer, listener)
		logs.Info("successful start-up of local tcp monitoring, port", l.Port)
		conn.Accept(listener, func(c net.Conn) {
			logs.Trace("new %s connection", l.Type)
			if l.Type == "secret" {
				handleSecret(c, config, l)
			} else if l.Type == "p2p" {
				handleP2PVisitor(c, config, l)
			}
		})
	}
	return nil
}
