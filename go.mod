module github.com/tsurubee/sshr

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Gurpartap/logrus-stack v0.0.0-20170710170904-89c00d8a28f4
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/lestrrat/go-server-starter v0.0.0-20180220115249-6ac0b358431b
	github.com/pkg/errors v0.8.1 // indirect
	github.com/pkg/sftp v1.10.0
	github.com/sirupsen/logrus v1.3.0
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20190219172222-a4c6cb3142f2
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sys v0.0.0-20190221075227-b4e8571b14e0 // indirect
)

replace golang.org/x/crypto => github.com/tsurubee/sshr.crypto v0.0.0-20181101225729-a944237b3cab
