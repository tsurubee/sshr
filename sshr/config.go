package sshr

import (
	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

const AuthTypePassword = "password"

type config struct {
	ListenAddr   string            `toml:"listen_addr"`
	RemoteAddr   string            `toml:"remote_addr"`
	AuthType     string            `toml:"auth_type"`
	ServerConfig *ssh.ServerConfig
	ClientConfig *ssh.ClientConfig
}

func loadConfig(path string) (*config, error) {
	var c config
	defaultConfig(&c)

	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}

	ctx := newContext(&c)
	ServerConfig, err := newServerConfig(&c, ctx)
	if err != nil {
		return nil, err
	}
	c.ServerConfig = ServerConfig

	return &c, nil
}

func newServerConfig(c *config, ctx *Context) (*ssh.ServerConfig, error) {
	serverConfig := &ssh.ServerConfig{}
	serverConfig.SetDefaults()

	if c.AuthType == AuthTypePassword {
		serverConfig.PasswordCallback =
			func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				ctx.Username = c.User()
				ctx.Password = string(pass)
				return nil, nil
			}
	}

	privateKeyBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return nil, err
	}
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	serverConfig.AddHostKey(privateKey)
	return serverConfig, nil
}

func defaultConfig(config *config) {
	config.ListenAddr = "0.0.0.0:2222"
}