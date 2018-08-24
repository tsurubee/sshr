// This file is implemented with reference to tg123/sshpiper.
// Ref: https://github.com/tg123/sshpiper/blob/master/vendor/golang.org/x/crypto/ssh/sshpiper.go
// Thanks to @tg123

package ssh

import (
	"errors"
	"fmt"
	"net"
)

type AuthPipeType int

const (
	// AuthPipeTypePassThrough does nothing but pass auth message to upstream
	AuthPipeTypePassThrough AuthPipeType = iota

	// AuthPipeTypeMap converts auth message to AuthMethod return by callback and pass it to upstream
	AuthPipeTypeMap

	// AuthPipeTypeDiscard discards auth message, do not pass it to upstream
	AuthPipeTypeDiscard

	// AuthPipeTypeNone converts auth message to NoneAuth and pass it to upstream
	AuthPipeTypeNone
)

// AuthPipe contains the callbacks of auth msg mapping from downstream to upstream
type AuthPipe struct {
	User                    string
	PasswordCallback        func(conn ConnMetadata, password []byte) (AuthPipeType, AuthMethod, error)
	PublicKeyCallback       func(conn ConnMetadata, key PublicKey)   (AuthPipeType, AuthMethod, error)
	UpstreamHostKeyCallback HostKeyCallback
}

type ProxyConfig struct {
	Config
	ServerConfig     *ServerConfig
	ClientConfig     *ClientConfig
	FindUpstreamHook func(username string) (string, error)
	Destination      string
	DestinationPort  string
	ServerVersion    string
}

// PipedConn provides downstream and upstream connections across proxy servers.
type PipedConn struct {
	Upstream          *connection
	Downstream        *connection
	upstreamMsgHook   func(msg []byte) ([]byte, error)
	downstreamMsgHook func(msg []byte) ([]byte, error)
}

// ProxyConn is a piped SSH connection, linking upstream ssh server and
// downstream ssh client together.
type ProxyConn struct {
	*PipedConn
	UpstreamMsgHook func(conn ConnMetadata, msg []byte) ([]byte, error)
	DownstreamHook  func(conn ConnMetadata, msg []byte) ([]byte, error)
}

func (p *ProxyConn) Wait() error {
	p.PipedConn.upstreamMsgHook = func(msg []byte) ([]byte, error) {
		if p.UpstreamMsgHook != nil {
			return p.UpstreamMsgHook(p.Downstream, msg)
		}

		return msg, nil
	}

	p.PipedConn.downstreamMsgHook = func(msg []byte) ([]byte, error) {
		if p.DownstreamHook != nil {
			return p.DownstreamHook(p.Downstream, msg)
		}

		return msg, nil
	}

	return p.PipedConn.loop()
}

// Close the piped connection create by SSHPiper
func (p *ProxyConn) Close() {
	p.PipedConn.Close()
}

func (pipe *PipedConn) processAuthMsg(msg *userAuthRequestMsg, authPipe *AuthPipe) (*userAuthRequestMsg, error) {

	var authType = AuthPipeTypePassThrough
	var authMethod AuthMethod
	mappedUser := authPipe.User

	switch msg.Method {
	case "publickey":
		if authPipe.PublicKeyCallback == nil {
			break
		}

		downKey, isQuery, sig, err := parsePublicKeyMsg(msg)
		if err != nil {
			return nil, err
		}

		authType, authMethod, err = authPipe.PublicKeyCallback(pipe.Downstream, downKey)
		if err != nil {
			return nil, err
		}

		if isQuery {
			// reply for query msg
			// skip query from upstream
			err = pipe.ack(downKey)
			if err != nil {
				return nil, err
			}

			// discard msg
			return nil, nil
		}

		ok, err := pipe.checkPublicKey(msg, downKey, sig)

		if err != nil {
			return nil, err
		}

		if !ok {
			return noneAuthMsg(mappedUser), nil
		}

	case "password":
		if authPipe.PasswordCallback == nil {
			break
		}

		payload := msg.Payload
		if len(payload) < 1 || payload[0] != 0 {
			return nil, parseError(msgUserAuthRequest)
		}
		payload = payload[1:]
		password, payload, ok := parseString(payload)
		if !ok || len(payload) > 0 {
			return nil, parseError(msgUserAuthRequest)
		}
		authType, authMethod, _ = authPipe.PasswordCallback(pipe.Downstream, password)

	default:
	}

	switch authType {
	case AuthPipeTypePassThrough:
		msg.User = mappedUser
		return msg, nil
	case AuthPipeTypeDiscard:
		return nil, nil
	case AuthPipeTypeNone:
		return noneAuthMsg(mappedUser), nil
	case AuthPipeTypeMap:
	}

	switch authMethod.method() {
	case "publickey":
		f, ok := authMethod.(publicKeyCallback)

		if !ok {
			return nil, errors.New("sshr: publicKeyCallback type assertions failed")
		}

		signers, err := f()
		// no mapped user change it to none or error occur
		if err != nil || len(signers) == 0 {
			return nil, err
		}

		for _, signer := range signers {
			msg, err = pipe.signAgain(mappedUser, msg, signer)
			if err != nil {
				return nil, err
			}
			return msg, nil
		}
	case "password":

		f, ok := authMethod.(passwordCallback)

		if !ok {
			return nil, errors.New("sshr: passwordCallback type assertions failed")
		}

		pw, err := f()
		if err != nil {
			return nil, err
		}

		type passwordAuthMsg struct {
			User     string `sshtype:"50"`
			Service  string
			Method   string
			Reply    bool
			Password string
		}

		Unmarshal(Marshal(passwordAuthMsg{
			User:     mappedUser,
			Service:  serviceSSH,
			Method:   "password",
			Reply:    false,
			Password: pw,
		}), msg)

		return msg, nil

	default:
	}

	msg.User = mappedUser
	return msg, nil
}

func (pipe *PipedConn) ack(key PublicKey) error {
	okMsg := userAuthPubKeyOkMsg {
		Algo:   key.Type(),
		PubKey: key.Marshal(),
	}

	return pipe.Downstream.transport.writePacket(Marshal(&okMsg))
}

func (pipe *PipedConn) checkPublicKey(msg *userAuthRequestMsg, pubkey PublicKey, sig *Signature) (bool, error) {
	if !isAcceptableAlgo(sig.Format) {
		return false, fmt.Errorf("ssh: algorithm %q not accepted", sig.Format)
	}
	signedData := buildDataSignedForAuth(pipe.Downstream.transport.getSessionID(), *msg, []byte(pubkey.Type()), pubkey.Marshal())

	if err := pubkey.Verify(signedData, sig); err != nil {
		return false, nil
	}

	return true, nil
}

func (pipe *PipedConn) signAgain(user string, msg *userAuthRequestMsg, signer Signer) (*userAuthRequestMsg, error) {
	rand      := pipe.Upstream.transport.config.Rand
	session   := pipe.Upstream.transport.getSessionID()
	upKey     := signer.PublicKey()
	upKeyData := upKey.Marshal()

	sign, err := signer.Sign(rand, buildDataSignedForAuth(session, userAuthRequestMsg{
		User:    user,
		Service: serviceSSH,
		Method:  "publickey",
	}, []byte(upKey.Type()), upKeyData))
	if err != nil {
		return nil, err
	}

	// manually wrap the serialized signature in a string
	s := Marshal(sign)
	sig := make([]byte, stringLength(len(s)))
	marshalString(sig, s)

	pubkeyMsg := &publickeyAuthMsg{
		User:     user,
		Service:  serviceSSH,
		Method:   "publickey",
		HasSig:   true,
		Algoname: upKey.Type(),
		PubKey:   upKeyData,
		Sig:      sig,
	}

	Unmarshal(Marshal(pubkeyMsg), msg)

	return msg, nil
}

func parsePublicKeyMsg(userAuthReq *userAuthRequestMsg) (PublicKey, bool, *Signature, error) {
	if userAuthReq.Method != "publickey" {
		return nil, false, nil, fmt.Errorf("not a publickey auth msg")
	}

	payload := userAuthReq.Payload
	if len(payload) < 1 {
		return nil, false, nil, parseError(msgUserAuthRequest)
	}
	isQuery := payload[0] == 0
	payload = payload[1:]
	algoBytes, payload, ok := parseString(payload)
	if !ok {
		return nil, false, nil, parseError(msgUserAuthRequest)
	}
	algo := string(algoBytes)
	if !isAcceptableAlgo(algo) {
		return nil, false, nil, fmt.Errorf("ssh: algorithm %q not accepted", algo)
	}

	pubKeyData, payload, ok := parseString(payload)
	if !ok {
		return nil, false, nil, parseError(msgUserAuthRequest)
	}

	pubKey, err := ParsePublicKey(pubKeyData)
	if err != nil {
		return nil, false, nil, err
	}

	var sig *Signature
	if !isQuery {
		sig, payload, ok = parseSignature(payload)
		if !ok || len(payload) > 0 {
			return nil, false, nil, parseError(msgUserAuthRequest)
		}
	}

	return pubKey, isQuery, sig, nil
}

func piping(dst, src packetConn, hooker func(msg []byte) ([]byte, error)) error {
	for {
		p, err := src.readPacket()

		if err != nil {
			return err
		}

		p, err = hooker(p)

		if err != nil {
			return err
		}

		err = dst.writePacket(p)

		if err != nil {
			return err
		}
	}
}

func (pipe *PipedConn) loop() error {
	c := make(chan error)

	go func() {
		c <- piping(pipe.Upstream.transport, pipe.Downstream.transport, pipe.downstreamMsgHook)
	}()

	go func() {
		c <- piping(pipe.Downstream.transport, pipe.Upstream.transport, pipe.upstreamMsgHook)
	}()

	defer pipe.Close()
	return <-c
}

func (pipe *PipedConn) Close() {
	pipe.Upstream.transport.Close()
	pipe.Downstream.transport.Close()
}

func (pipe *PipedConn) pipeAuthSkipBanner(packet []byte) (bool, error) {
	err := pipe.Upstream.transport.writePacket(packet)
	if err != nil {
		return false, err
	}

	for {
		packet, err := pipe.Upstream.transport.readPacket()
		if err != nil {
			return false, err
		}

		msgType := packet[0]

		if err = pipe.Downstream.transport.writePacket(packet); err != nil {
			return false, err
		}

		switch msgType {
		case msgUserAuthSuccess:
			return true, nil
		case msgUserAuthBanner:
			// should read another packet from upstream
			continue
		case msgUserAuthFailure:
		default:
		}

		return false, nil
	}
}

func (pipe *PipedConn) PipeAuth(initUserAuthMsg *userAuthRequestMsg, authPipe *AuthPipe) error {
	err := pipe.Upstream.sendAuthReq()
	if err != nil {
		return err
	}

	userAuthMsg := initUserAuthMsg

	for {
		userAuthMsg, err = pipe.processAuthMsg(userAuthMsg, authPipe)
		if err != nil {
			return err
		}

		if userAuthMsg != nil {
			succ, err := pipe.pipeAuthSkipBanner(Marshal(userAuthMsg))
			if err != nil {
				return err
			}
			if succ {
				return nil
			}
		}

		var packet []byte

		for {
			// find next msg which need to be hooked
			if packet, err = pipe.Downstream.transport.readPacket(); err != nil {
				return err
			}

			// we can only handle auth req at the moment
			if packet[0] == msgUserAuthRequest {
				// should hook, deal with it
				break
			}

			// pipe other auth msg
			succ, err := pipe.pipeAuthSkipBanner(packet)
			if err != nil {
				return err
			}
			if succ {
				return nil
			}
		}

		var userAuthReq userAuthRequestMsg

		if err = Unmarshal(packet, &userAuthReq); err != nil {
			return err
		}

		userAuthMsg = &userAuthReq
	}
}

func (c *connection) sendAuthReq() error {
	if err := c.transport.writePacket(Marshal(&serviceRequestMsg{serviceUserAuth})); err != nil {
		return err
	}

	packet, err := c.transport.readPacket()
	if err != nil {
		return err
	}
	var serviceAccept serviceAcceptMsg
	return Unmarshal(packet, &serviceAccept)
}

func NewDownstream(c net.Conn, config *ServerConfig) (*connection, error) {
	fullConf := *config
	fullConf.SetDefaults()

	s := &connection{
		sshConn: sshConn{conn: c},
	}

	_, err := s.serverHandshakeNoAuth(&fullConf)
	if err != nil {
		c.Close()
		return nil, err
	}

	return s, nil
}

func NewUpstream(c net.Conn, addr string, config *ClientConfig) (*connection, error) {
	fullConf := *config
	fullConf.SetDefaults()

	conn := &connection{
		sshConn: sshConn{conn: c},
	}

	if err := conn.clientHandshakeNoAuth(addr, &fullConf); err != nil {
		c.Close()
		return nil, err
	}

	return conn, nil
}

func (c *connection) NextAuthMsg() (*userAuthRequestMsg, error) {
	var userAuthReq userAuthRequestMsg

	if packet, err := c.transport.readPacket(); err != nil {
		return nil, err
	} else if err = Unmarshal(packet, &userAuthReq); err != nil {
		return nil, err
	}

	if userAuthReq.Service != serviceSSH {
		return nil, errors.New("ssh: client attempted to negotiate for unknown service: " + userAuthReq.Service)
	}

	return &userAuthReq, nil
}

func noneAuthMsg(user string) *userAuthRequestMsg {
	return &userAuthRequestMsg{
		User:    user,
		Service: serviceSSH,
		Method:  "none",
	}
}

func (c *connection) clientHandshakeNoAuth(dialAddress string, config *ClientConfig) error {
	c.clientVersion = []byte(packageVersion)
	if config.ClientVersion != "" {
		c.clientVersion = []byte(config.ClientVersion)
	}

	var err error
	c.serverVersion, err = exchangeVersions(c.sshConn.conn, c.clientVersion)
	if err != nil {
		return err
	}

	c.transport = newClientTransport(
		newTransport(c.sshConn.conn, config.Rand, true /* is client */),
		c.clientVersion, c.serverVersion, config, dialAddress, c.sshConn.RemoteAddr())

	if err := c.transport.waitSession(); err != nil {
		return err
	}

	c.sessionID = c.transport.getSessionID()
	return nil
}

func (c *connection) serverHandshakeNoAuth(config *ServerConfig) (*Permissions, error) {
	if len(config.hostKeys) == 0 {
		return nil, errors.New("ssh: server has no host keys")
	}

	var err error
	if config.ServerVersion != "" {
		c.serverVersion = []byte(config.ServerVersion)
	} else {
		c.serverVersion = []byte("SSH-2.0-Go")
	}
	c.clientVersion, err = exchangeVersions(c.sshConn.conn, c.serverVersion)
	if err != nil {
		return nil, err
	}

	tr := newTransport(c.sshConn.conn, config.Rand, false /* not client */)
	c.transport = newServerTransport(tr, c.clientVersion, c.serverVersion, config)

	if err := c.transport.waitSession(); err != nil {
		return nil, err

	}
	c.sessionID = c.transport.getSessionID()

	var packet []byte
	if packet, err = c.transport.readPacket(); err != nil {
		return nil, err
	}

	var serviceRequest serviceRequestMsg
	if err = Unmarshal(packet, &serviceRequest); err != nil {
		return nil, err
	}
	if serviceRequest.Service != serviceUserAuth {
		return nil, errors.New("ssh: requested service '" + serviceRequest.Service + "' before authenticating")
	}
	serviceAccept := serviceAcceptMsg{
		Service: serviceUserAuth,
	}
	if err := c.transport.writePacket(Marshal(&serviceAccept)); err != nil {
		return nil, err
	}

	return nil, nil
}

