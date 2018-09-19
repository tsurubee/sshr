package main

import (
	"flag"
	"testing"
	"os"
	"fmt"
	"golang.org/x/crypto/ssh"
)

var (
	integration = flag.Bool("integration", false, "run integration tests")
)

func loginByPassword(port int, t *testing.T) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User: "tsurubee",
		Auth: []ssh.AuthMethod{ssh.Password("test")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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
	client, _, err := loginByPassword(2222, t)
	if err != nil {
		t.Errorf("integration.TestLoginByPassword() error = %v, wantErr %v", err, nil)
	}
	defer client.Close()
}