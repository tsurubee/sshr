package sshr

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type logger struct {
	user string
}

func (l *logger) infof(format string, args ...interface{}) {
	format = fmt.Sprintf("[user:%s] %s", l.user, format)
	logrus.Infof(format, args...)
}
