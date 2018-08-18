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

	serverConfig, err := newServerConfig()
	proxy.ServerConfig = serverConfig

	return &SSHServer{
		config:      c,
		ProxyConfig: proxy,
	}, nil
}

func (server *SSHServer) Listen() (err error) {
	server.listener, err = net.Listen("tcp", server.config.ListenAddr)
	if err != nil {
		return err
	}

	logrus.Info("Start Listening...")
	return err
}

func (server *SSHServer) Serve() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if server.listener != nil {
				return err
			}
		}
		logrus.Info("SSH Client connected ", "clientIp ", conn.RemoteAddr())

		go func() {
			p, err := ssh.NewSSHProxyConn(conn, server.ProxyConfig)
			if err != nil {
				logrus.Fatal(err)
				return
			}

			if err = p.Wait(); err != nil {
				logrus.Fatal(err)
				return
			}
		}()
	}
}

func (server *SSHServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	return server.Serve()
}