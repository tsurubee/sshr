# sshr
sshr is an SSH proxy server whose client is not aware of the connection destination.  
A developer using sshr can freely incorporate own hooks that dynamically determines the destination host from the SSH username.  
Therefore, for example, when the server administrator wants to centrally manage the linkage information between the SSH user and the server to be used in the DB, you can refer to the destination host from the DB with the hook that can be pluggable in sshr.

<img src="./docs/images/conceptual_scheme.png" alt="conceptual_scheme" width="700">

## Usage

## Installation
```
$ go get github.com/tsurubee/ssht
```

## License

[MIT](https://github.com/tsurubee/sshr/blob/master/LICENSE)

## Author

[tsurubee](https://github.com/tsurubee)
