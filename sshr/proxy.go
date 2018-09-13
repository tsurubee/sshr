package sshr

import (
	"net"
	"golang.org/x/crypto/ssh"
)

func newSSHProxyConn(conn net.Conn, proxyConf *ssh.ProxyConfig) (proxyConn *ssh.ProxyConn, err error) {
	d, err := ssh.NewDownstreamConn(conn, proxyConf.ServerConfig)
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
	proxyConf.User = username
	upstreamHost, err := proxyConf.FindUpstreamHook(username)
	if err != nil {
		return nil, err
	}
	proxyConf.DestinationHost = upstreamHost

	upConn, err := net.Dial("tcp", proxyConf.DestinationHost + ":" + proxyConf.DestinationPort)
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

	if err = p.AuthenticateProxyConn(authRequestMsg, proxyConf); err != nil {
		return nil, err
	}

	return p, nil
}
