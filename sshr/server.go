package sshr

import (
	"net"
	"github.com/sirupsen/logrus"
)

type SSHServer struct {
	listener net.Listener
	config   *config
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

func (server * SSHServer) Listen() (err error) {
	server.listener, err = net.Listen("tcp", server.config.ListenAddr)
	if err != nil {
		return err
	}

	logrus.Info("Listening address ", server.listener.Addr())
	return err
}

func (server * SSHServer) Serve() error {
	logrus.Info("Serve")
	return nil
}

func (server * SSHServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	return server.Serve()
}