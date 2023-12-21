module github.com/tsurubee/sshr

go 1.21.4

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/lestrrat/go-server-starter v0.0.0-20180220115249-6ac0b358431b
	github.com/pkg/sftp v1.13.6
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/crypto v0.17.0
	golang.org/x/sync v0.5.0
)

require (
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

replace golang.org/x/crypto => github.com/tsurubee/sshr.crypto v0.0.0-20231220131018-9dc964a49cf7
