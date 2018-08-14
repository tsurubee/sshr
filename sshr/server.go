package sshr

import (
	"net"
	"github.com/sirupsen/logrus"
)

type AuthenticationHook func(*Context, string) error

type SSHServer struct {
	listener           net.Listener
	config             *config
	AuthenticationHook AuthenticationHook
}

func NewSSHServer(confFile string) (*SSHServer, error) {
	c, err := loadConfig(confFile)
	if err != nil {
		return nil, err
	}
	return &SSHServer{
		config: c,
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
			if err := NewSSHPipeConn(conn, server.config); err != nil {
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