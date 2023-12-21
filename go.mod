module github.com/tsurubee/sshr

go 1.20

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/lestrrat/go-server-starter v0.0.0-20180220115249-6ac0b358431b
	github.com/pkg/sftp v1.10.0
	github.com/sirupsen/logrus v1.9.2
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550
	golang.org/x/sync v0.2.0
)

require (
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/sys v0.8.0 // indirect
)

replace golang.org/x/crypto => github.com/tsurubee/sshr.crypto v0.0.0-20230523043435-1b27928394e9
