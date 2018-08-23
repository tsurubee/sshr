package sshr

import (
	"net"
	"golang.org/x/crypto/ssh"
)

func NewSSHProxyConn(conn net.Conn, proxy *ssh.ProxyConfig) (pipe *ssh.ProxyConn, err error) {
	d, err := ssh.NewDownstream(conn, proxy.ServerConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if pipe == nil {
			d.Close()
		}
	}()

	userAuthReq, err := d.NextAuthMsg()
	if err != nil {
		return nil, err
	}

	username := userAuthReq.User

	upstream_host, err := proxy.FindUpstreamHook(username)
	if err != nil {
		return nil, err
	}

	upconn, err := net.Dial("tcp", upstream_host + ":" + proxy.DestinationPort)
	if err != nil {
		return nil, err
	}

	mappedUser := username
	authPipe := &ssh.AuthPipe{
		User: mappedUser,
		UpstreamHostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := upconn.RemoteAddr().String()

	u, err := ssh.NewUpstream(upconn, addr, &ssh.ClientConfig{
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
