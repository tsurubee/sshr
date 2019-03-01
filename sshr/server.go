package sshr

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/sync/errgroup"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"github.com/lestrrat/go-server-starter/listener"
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
		conn = conn.(*net.TCPConn)
		if err != nil {
			if os.Getenv("SERVER_STARTER_PORT") != "" {
				break
			}

			if server.listener != nil {
				return err
			}
		}
		logrus.Info("SSH Client connected. ", "ClientIP=", conn.RemoteAddr())

		eg.Go(func() error {
			p, err := newSSHProxyConn(conn, server.ProxyConfig)
			if err != nil {
				logrus.Infof("Connection from %v closed. %v", conn.RemoteAddr(), err)
				return err
			}
			logrus.Infof("Establish a proxy connection between %v and %v", conn.RemoteAddr(), server.ProxyConfig.DestinationHost)
			err = p.Wait()
			logrus.Infof("Connection from %v closed. %v", conn.RemoteAddr(), err)
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

	ch := make(chan os.Signal)
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
