package openapi

import (
	"strings"

	"net/http"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func getIP(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if "" != ip {
		return strings.Split(ip, ",")[0]
	}
	return strings.Split(req.RemoteAddr, ":")[0]
}

func parseValidateDirective(oapi *OpenAPI, d *caddyfile.Dispenser) error {

	args := d.RemainingArgs()
	if len(args) <= 0 {
		return d.ArgErr()
	}

	for _, token := range args {

		switch token {
		case VALUE_REQ_PARAMS:
			oapi.RequestParams = true

		default:
			return d.Errf("unrecognized validate option: '%s'", token)
		}
	}

	for nest := d.Nesting(); d.NextBlock(nest); {
		token := d.Val()
		switch token {
		}
	}

	return nil
}

//err := parseValidate(oapi, d)
