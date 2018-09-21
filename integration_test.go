package main

import (
	"flag"
	"testing"
	"os"
	"fmt"
	"io/ioutil"
	"time"
	"golang.org/x/crypto/ssh"
)

var (
	integration = flag.Bool("integration", false, "run integration tests")
)

func loginByPassword(port int, password string) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            "tsurubee",
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", port), sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}

func loginByPublicKey(port int, keyPath string) (*ssh.Client, *ssh.Session, error) {
	privateKeyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            "tsurubee",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(privateKey)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", port), sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestLoginByPassword(t *testing.T) {
	if !*integration {
		t.Skip()
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "success login",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "success login",
			password: "failpass",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, err := loginByPassword(2222, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("integration.TestLoginByPassword() error = %v, wantErr %v", err, nil)
				return
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestLoginByPublicKey(t *testing.T) {
	if !*integration {
		t.Skip()
	}

	tests := []struct {
		name    string
		keyPath string
		wantErr bool
	}{
		{
			name:    "success login",
			keyPath: "misc/testdata/client_keys/id_rsa",
			wantErr: false,
		},
		{
			name:    "incorrect SSH key",
			keyPath: "misc/testdata/client_keys/id_rsa_dummy",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, err := loginByPublicKey(2222, tt.keyPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("integration.TestLoginByPublicKey() error = %v, wantErr %v", err, nil)
				return
			}
			if client != nil {
				client.Close()
			}
		})
	}
}