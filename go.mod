module github.com/tsurubee/sshr

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052
	github.com/konsorten/go-windows-terminal-sequences v1.0.1
	github.com/kr/fs v0.1.0
	github.com/lestrrat/go-server-starter v0.0.0-20180220115249-6ac0b358431b
	github.com/pkg/errors v0.8.0
	github.com/pkg/sftp v1.8.3
	github.com/sirupsen/logrus v1.1.1
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20181019084534-8f1d3d21f81b
)

replace golang.org/x/crypto => github.com/tsurubee/sshr.crypto v0.0.0-20180922185822-82c4f2725339
