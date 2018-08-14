package sshr


type Context struct {
	UpstreamHost string
	Username     string
	Password     string
}

func newContext(c *config) *Context {
	return &Context{
	}
}
