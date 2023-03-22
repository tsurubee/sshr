package sshr

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lestrrat/go-server-starter/listener"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

type SSHServer struct {
	listener    net.Listener
	config      *config
	ProxyConfig *ssh.ProxyConfig
	shutdown    bool
}

func NewSSHServer(confFile string) (*SSHServer, error) {
	c, err := loadConfig(confFile)
	if err != nil {
		return nil, err
	}

	proxy := &ssh.ProxyConfig{}
	proxy.Config.SetDefaults()
	proxy.DestinationPort = c.DestinationPort
	proxy.UseMasterKey = c.UseMasterKey
	proxy.MasterKeyPath = c.MasterKeyPath

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
	if os.Getenv("SERVER_STARTER_PORT") != "" {
		listeners, err := listener.ListenAll()
		if listeners == nil || err != nil {
			return err
		}
		server.listener = listeners[0]
	} else {
		server.listener, err = net.Listen("tcp", server.config.ListenAddr)
		if err != nil {
			return err
		}
	}

	logrus.Info("Start Listening on ", server.listener.Addr())
	return err
}

func (server *SSHServer) serve() error {
	eg := errgroup.Group{}

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if os.Getenv("SERVER_STARTER_PORT") != "" {
				break
			}

			if server.listener != nil {
				return err
			}
		}
		tcpConn := conn.(*net.TCPConn)
		tcpConn.SetKeepAlive(true)
		logrus.Info("SSH Client connected. ", "ClientIP=", tcpConn.RemoteAddr())

		eg.Go(func() error {
			logger := &logger{
				user: "unknown",
			}
			p, err := newSSHProxyConn(tcpConn, server.ProxyConfig)
			if p != nil {
				logger.user = p.User
			}
			if err != nil {
				logger.info("Connection from %s closed. %v", tcpConn.RemoteAddr().String(), err)
				return err
			}
			logger.info("Establish a proxy connection between %s and %s with username %s", tcpConn.RemoteAddr().String(), p.DestinationHost, p.User)
			err = p.Wait()
			logger.info("Connection from %s closed.", tcpConn.RemoteAddr().String())
			return err
		})
	}

	return eg.Wait()
}

func (server *SSHServer) Run() error {
	var lastError error
	done := make(chan struct{})

	if err := server.listen(); err != nil {
		return err
	}

	go func() {
		if err := server.serve(); err != nil {
			if !server.shutdown {
				lastError = err
			}
		}
		done <- struct{}{}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM)
Loop:
	for {
		switch <-ch {
		case syscall.SIGHUP, syscall.SIGTERM:
			if err := server.stop(); err != nil {
				lastError = err
			}
			break Loop
		}
	}

	<-done
	return lastError
}

func (server *SSHServer) stop() error {
	server.shutdown = true
	if server.listener != nil {
		logrus.Info("Close listener")
		if err := server.listener.Close(); err != nil {
			return err
		}
	}
	return nil
}
