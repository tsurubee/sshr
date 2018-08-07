package sshr

import (
	"net"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
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

	logrus.Info("Start Listening...")
	return err
}

func (server * SSHServer) Serve() error {
	serverConfig := &ssh.ServerConfig{
		// ToDo
		// PasswordCallback
		// PublicKeyCallback
		NoClientAuth: true,
	}
	privateKeyBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}
	serverConfig.AddHostKey(privateKey)

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if server.listener != nil {
				return err
			}
		}
		logrus.Info("SSH Client connected ", "clientIp ", conn.RemoteAddr())

		// ToDo goroutine
	}
}

func (server * SSHServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	return server.Serve()
}