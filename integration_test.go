package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	integration = flag.Bool("integration", false, "run integration tests")
)

func loginByPassword(username string, port int, password string) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            username,
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

func loginByPublicKey(username string, port int, keyPath string) (*ssh.Client, *ssh.Session, error) {
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            username,
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

func execCommand(sess *ssh.Session, command string) (string, error) {
	output, err := sess.Output(command)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(output), "\n"), nil
}

func uploadFileByScp(sess *ssh.Session, uploadFile string, permission string) error {
	f, err := os.Open(uploadFile)
	if err != nil {
		return err
	}
	defer f.Close()
	filename := path.Base(uploadFile)

	contentsBytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	r := bytes.NewReader(contentsBytes)

	go func() {
		w, _ := sess.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C"+permission, int64(len(contentsBytes)), filename)
		io.Copy(w, r)
		fmt.Fprint(w, "\x00")
	}()

	return sess.Run("scp -tr ./")
}

func uploadFileBySftp(username string, port int, password string, uploadFile string) error {
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%d", port), sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	dstFile, err := client.Create(path.Base(uploadFile))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := os.Open(uploadFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
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
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "success login",
			username: "tsurubee",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "success login",
			username: "tsurubee",
			password: "failpass",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, err := loginByPassword(tt.username, 2222, tt.password)
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
			client, _, err := loginByPublicKey("tsurubee", 2222, tt.keyPath)
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

func TestExecHostnameCommand(t *testing.T) {
	if !*integration {
		t.Skip()
	}

	tests := []struct {
		name     string
		username string
		password string
		hostname string
		wantErr  bool
	}{
		{
			name:     "Get Hostname:tsurubee",
			username: "tsurubee",
			password: "testpass",
			hostname: "host-tsurubee",
			wantErr:  false,
		},
		{
			name:     "Get Hostname:hoge",
			username: "hoge",
			password: "testpass",
			hostname: "host-hoge",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sess, err := loginByPassword(tt.username, 2222, tt.password)
			if err != nil != tt.wantErr {
				t.Errorf("integration.TestExecHostnameCommand() error = %v, wantErr %v", err, nil)
				return
			}

			hostname, err := execCommand(sess, "hostname")
			if (err != nil) != tt.wantErr {
				t.Errorf("integration.TestExecHostnameCommand() error = %v, wantErr %v", err, nil)
				return
			}
			if hostname != tt.hostname {
				t.Errorf("integration.TestExecHostnameCommand() error = %v, wantErr %v", err, nil)
				return
			}
		})
	}
}

func TestUploadFileByScp(t *testing.T) {
	if !*integration {
		t.Skip()
	}

	tests := []struct {
		name       string
		username   string
		password   string
		uploadFile string
		wantErr    bool
	}{
		{
			name:       "success upload",
			username:   "tsurubee",
			password:   "testpass",
			uploadFile: "misc/testdata/upload_test/uploadByScp.txt",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sess, err := loginByPassword(tt.username, 2222, tt.password)
			if err != nil != tt.wantErr {
				t.Errorf("integration.TestUploadFileByScp() error = %v, wantErr %v", err, nil)
				return
			}

			err = uploadFileByScp(sess, tt.uploadFile, "0644")
			if (err != nil) != tt.wantErr {
				t.Errorf("integration.TestUploadFileByScp() error = %v, wantErr %v", err, nil)
				return
			}
		})
	}
}

func TestUploadFileBySftp(t *testing.T) {
	if !*integration {
		t.Skip()
	}

	tests := []struct {
		name       string
		username   string
		password   string
		uploadFile string
		wantErr    bool
	}{
		{
			name:       "success upload",
			username:   "tsurubee",
			password:   "testpass",
			uploadFile: "misc/testdata/upload_test/uploadBySftp.txt",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uploadFileBySftp(tt.username, 2222, tt.password, tt.uploadFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("integration.TestUploadFileByScp() error = %v, wantErr %v", err, nil)
				return
			}
		})
	}
}
