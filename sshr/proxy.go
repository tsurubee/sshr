package sshr

import (
	"net"
	"golang.org/x/crypto/ssh"
)

func NewSSHProxyConn(conn net.Conn, proxy *ssh.ProxyConfig) (pipe *ssh.ProxyConn, err error) {
	d, err := ssh.NewDownstreamConn(conn, proxy.ServerConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if pipe == nil {
			d.Close()
		}
	}()

	userAuthReq, err := d.GetAuthRequestMsg()
	if err != nil {
		return nil, err
	}

	username := userAuthReq.User
	upstreamHost, err := proxy.FindUpstreamHook(username)
	if err != nil {
		return nil, err
	}
	proxy.Destination = upstreamHost

	upconn, err := net.Dial("tcp", upstreamHost + ":" + proxy.DestinationPort)
	if err != nil {
		return nil, err
	}

	authPipe := &ssh.AuthPipe{
		User: username,
		UpstreamHostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr   := upconn.RemoteAddr().String()
	u, err := ssh.NewUpstreamConn(upconn, addr, &ssh.ClientConfig{
		HostKeyCallback: authPipe.UpstreamHostKeyCallback,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if pipe == nil {
			u.Close()
		}
	}()

	p := &ssh.PipedConn{
		Upstream:   u,
		Downstream: d,
	}

	if err = p.PipeAuth(userAuthReq, authPipe); err != nil {
		return nil, err
	}

	return &ssh.ProxyConn{PipedConn: p}, nil
}
