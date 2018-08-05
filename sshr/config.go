package sshr

import (
	"github.com/BurntSushi/toml"
)

type config struct {
	ListenAddr string `toml:"listen_addr"`
	RemoteAddr string `toml:"remote_addr"`
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

func defaultConfig(config *config) {
	config.ListenAddr = "0.0.0.0:2222"
}