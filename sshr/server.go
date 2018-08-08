package sshr

import (
	"net"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
)

type middlewareFunc func(*Context, string) error
type middleware map[string]middlewareFunc

type SSHServer struct {
	listener   net.Listener
	config     *config
	middleware middleware
}

func NewSSHServer(confFile string) (*SSHServer, error) {
	c, err := loadConfig(confFile)
	if err != nil {
		return nil, err
	}
	m := middleware{}
	return &SSHServer{
		config:     c,
		middleware: m,
	}, nil
}

func (server *SSHServer) Use(command string, m middlewareFunc) {
	server.middleware[strings.ToUpper(command)] = m
}

func getUsername(sshConn *ssh.ServerConn) (string, error) {
	return "tsurubee", nil
}

func getPassword(sshConn *ssh.ServerConn) (string, error) {
	return "password", nil
}

func startSSHProxy(conn net.Conn, c *config) error {
	sshConn, _, _, err := ssh.NewServerConn(conn, c.ServerConfig)
	if err != nil {
		return err
	}
	// Get username and password from SSH session
	username, err := getUsername(sshConn)
	password, err := getPassword(sshConn)
	if err != nil {
		return err
	}

	// client
	context := newContext(c)
	if err := FindUpstreamByUsername(context, username); err != nil {
		return err
	}
	ClientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			//ssh.PublicKeys("keys"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", context.UpstreamHost + ":22", ClientConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	sess, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	//ToDo
	// pty

	//ToDo
	//client <-> proxy <-> upstream
	//dual-directional io.Copy
	return nil
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

		go startSSHProxy(conn, server.config)
	}
}

func (server *SSHServer) ListenAndServe() error {
	if err := server.Listen(); err != nil {
		return err
	}

	return server.Serve()
}