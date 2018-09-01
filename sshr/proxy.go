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

	authRequestMsg, err := d.GetAuthRequestMsg()
	if err != nil {
		return nil, err
	}

	username := authRequestMsg.User
	proxy.User = username
	upstreamHost, err := proxy.FindUpstreamHook(username)
	if err != nil {
		return nil, err
	}
	proxy.DestinationHost = upstreamHost

	upConn, err := net.Dial("tcp", proxy.DestinationHost + ":" + proxy.DestinationPort)
	if err != nil {
		return nil, err
	}

	u, err := ssh.NewUpstreamConn(upConn, &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if pipe == nil {
			u.Close()
		}
	}()

	p := &ssh.ProxyConn{
		Upstream:   u,
		Downstream: d,
	}

	if err = p.ProxyAuthenticate(authRequestMsg, proxy); err != nil {
		return nil, err
	}

	return p, nil
}
