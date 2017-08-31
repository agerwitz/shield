package api

import (
	"net/http"
	"net/url"
	"strings"
)

func Unlock(master string) error {
	uri, err := ShieldURI("/v2/unlock")
	if err != nil {
		return err
	}

	respMap := make(map[string]string)
	r, err := http.NewRequest("POST", uri.String(), strings.NewReader(url.Values{"master_password": {master}}.Encode()))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return uri.Request(&respMap, r)
}
