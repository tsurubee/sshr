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
	p := &ssh.ProxyConn{
		User:       username,
		Downstream: d,
	}
	upstreamHost, err := proxyConf.FindUpstreamHook(username)
	if err != nil {
		if err := p.SendFailureMsg(err.Error()); err != nil {
			return p, err
		}
		return p, err
	}
	p.DestinationHost = upstreamHost

	upConn, err := net.Dial("tcp", upstreamHost+":"+proxyConf.DestinationPort)
	if err != nil {
		return p, err
	}

	u, err := ssh.NewUpstreamConn(upConn, &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return p, err
	}
	defer func() {
		if proxyConn == nil {
			u.Close()
		}
	}()

	p.Upstream = u

	if err = p.AuthenticateProxyConn(authRequestMsg, proxyConf); err != nil {
		return p, err
	}

	return p, nil
}
