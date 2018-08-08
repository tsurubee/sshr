package sshr

type Context struct {
	RemoteAddr   string
	UpstreamHost string
	Username     string
	Password     string
}

func newContext(c *config) *Context {
	return &Context{
		RemoteAddr: c.RemoteAddr,
	}
}
