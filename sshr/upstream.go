package sshr

import (
	"errors"
)

func FindUpstreamByUsername(c *Context, username string) error {
	// ToDo: Find upstream host from RESTful API
	if username == "tsurubee" {
		c.UpstreamHost = "host-tsurubee"
		return nil
	} else {
		return errors.New(username + "'s host is not found!")
	}
}
