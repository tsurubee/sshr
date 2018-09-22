package sshr

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c, err := loadConfig("../example.toml")
	if err != nil {
		t.Errorf("Config load failed: %v", err)
	}

	if c.ListenAddr != "0.0.0.0:2222" {
		t.Errorf("Config cannot parse ListenAddr")
	}
	if c.RemoteAddr != "127.0.0.1:2222" {
		t.Errorf("Config cannot parse RemoteAddr")
	}
	if c.DestinationPort != "22" {
		t.Errorf("Config cannot parse DestinationPort")
	}
	if c.HostKeyPath != "misc/testdata/hostkey/id_rsa" {
		t.Errorf("Config cannot parse HostKeyPath")
	}
}

func TestNewServerConfig(t *testing.T) {
	c, err := loadConfig("../example.toml")
	if err != nil {
		t.Errorf("Config load failed: %v", err)
	}
	c.HostKeyPath = "../misc/testdata/hostkey/id_rsa"

	serverConfig, err := newServerConfig(c)
	if err != nil || serverConfig == nil {
		t.Errorf("ServerConfig cannot be set: %v", err)
	}
}
