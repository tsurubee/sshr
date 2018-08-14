package sshr

import (
	"net"
	"golang.org/x/crypto/ssh"
)

func NewSSHPipeConn(conn net.Conn, c *config) error {
	_, _, _, err := ssh.NewServerConn(conn, c.ServerConfig)
	if err != nil {
		return err
	}
	return nil
}

