package sshr

import (
	"net"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHServer struct {
	listener    net.Listener
	config      *config
	ProxyConfig *ssh.ProxyConfig
}

func NewSSHServer(confFile string) (*SSHServer, error) {
	c, err := loadConfig(confFile)
	if err != nil {
		return nil, err
	}

	proxy := &ssh.ProxyConfig{}
	proxy.Config.SetDefaults()
	proxy.DestinationPort = c.DestinationPort

	serverConfig, err := newServerConfig(c)
	if err != nil {
		return nil, err
	}
	proxy.ServerConfig = serverConfig

	return &SSHServer{
		config:      c,
		ProxyConfig: proxy,
	}, nil
}

func (server *SSHServer) listen() (err error) {
	server.listener, err = net.Listen("tcp", server.config.ListenAddr)
	if err != nil {
		return err
	}

	logrus.Info("Start Listening on ", server.listener.Addr())
	return err
}

func (server *SSHServer) serve() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if server.listener != nil {
				return err
			}
		}
		logrus.Info("SSH Client connected. ", "ClientIP=", conn.RemoteAddr())

		go func() {
			p, err := newSSHProxyConn(conn, server.ProxyConfig)
			if err != nil {
				logrus.Infof("Connection from %v closed. %v", conn.RemoteAddr(), err)
				return
			}
			logrus.Infof("Establish a proxy connection between %v and %v", conn.RemoteAddr(), server.ProxyConfig.DestinationHost)
			err = p.Wait()
			logrus.Infof("Connection from %v closed. %v", conn.RemoteAddr(), err)
		}()
	}
}

func (server *SSHServer) ListenAndServe() error {
	if err := server.listen(); err != nil {
		return err
	}

	return server.serve()
}