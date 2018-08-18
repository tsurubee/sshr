package sshr

import (
	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

type config struct {
	ListenAddr string `toml:"listen_addr"`
	RemoteAddr string `toml:"remote_addr"`
	AuthType   string `toml:"auth_type"`
}

func loadConfig(path string) (*config, error) {
	var c config
	defaultConfig(&c)

	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func newServerConfig() (*ssh.ServerConfig, error) {
	serverConfig := &ssh.ServerConfig{}

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