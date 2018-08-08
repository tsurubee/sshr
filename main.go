package main

import (
	"github.com/sirupsen/logrus"
	"github.com/Gurpartap/logrus-stack"
	"github.com/tsurubee/sshr/sshr"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel}
	logrus.AddHook(logrus_stack.NewHook(stackLevels, stackLevels))
}

func main() {
	confFile := "./example.toml"

	sshServer, err := sshr.NewSSHServer(confFile)
	if err != nil {
		logrus.Fatal(err)
	}

	sshServer.Use("findUpstream", sshr.FindUpstreamByUsername)
	if err := sshServer.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}