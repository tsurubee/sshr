# sshr
[![Build Status](https://travis-ci.org/tsurubee/sshr.svg?branch=master)](https://travis-ci.org/tsurubee/sshr)  

sshr is an SSH proxy server whose client is not aware of the connection destination.  
A developer using sshr can freely incorporate own hooks that dynamically determines the destination host from the SSH username.  Therefore, for example, when the server administrator wants to centrally manage the linkage information between the SSH user and the server to be used in the DB, you can refer to the destination host from the DB with the hook that can be pluggable in sshr.

<img src="./docs/images/conceptual_scheme.png" alt="conceptual_scheme" width="800">

## Usage
### Installation
```
$ go get github.com/tsurubee/sshr
```

### Example
```go
func main() {
	confFile := "./example.toml"

	sshServer, err := sshr.NewSSHServer(confFile)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	sshServer.ProxyConfig.FindUpstreamHook = FindUpstreamByUsername
	if err := sshServer.ListenAndServe(); err != nil {
		fmt.Errorf("Error: %s", err)
	}
}

func FindUpstreamByUsername(username string) (string, error) {
	if username == "tsurubee" {
		return "host-tsurubee", nil
	} else {
		return "", errors.New(username + "'s host is not found!")
	}
}

```
FindUpstreamHook is a hook that can be pluggable, which allows you to write your own logic to dynamically determine the destination host from the SSH username.

## License

[MIT](https://github.com/tsurubee/sshr/blob/master/LICENSE)

## Author

[tsurubee](https://github.com/tsurubee)
