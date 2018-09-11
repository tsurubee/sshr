package sshr

import (
	"net"
	"golang.org/x/crypto/ssh"
)

func newSSHProxyConn(conn net.Conn, proxy *ssh.ProxyConfig) (proxyConn *ssh.ProxyConn, err error) {
	d, err := ssh.NewDownstreamConn(conn, proxy.ServerConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if proxyConn == nil {
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
		if proxyConn == nil {
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
