package sshr

import "errors"

func findUpstreamByUsername(username string) (string, error) {
	if username == "tsurubee" {
		return "host-tsurubee", nil
	} else  {
		return "", errors.New(username + "'s host is not found!")
	}
}
