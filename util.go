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

func parseCheckDirective(oapi *OpenAPI, d *caddyfile.Dispenser) error {

	args := d.RemainingArgs()
	if len(args) != 0 {
		return d.ArgErr()
	}

	oapi.Check = &CheckOptions{RequestBody: false, RequestParams: false}

	/*
		for _, token := range args {

			switch token {
			case VALUE_REQ_PARAMS:
				oapi.Check.RequestParams = true

			case VALUE_REQ_BODY:
				oapi.Check.RequestParams = true
				oapi.Check.RequestBody = true

			default:
				return d.Errf("unrecognized validate option: '%s'", token)
			}
		}
	*/

	for nest := d.Nesting(); d.NextBlock(nest); {
		token := d.Val()
		switch token {
		case VALUE_REQ_PARAMS:
			oapi.Check.RequestParams = true

		case VALUE_REQ_BODY:
			oapi.Check.RequestParams = true
			oapi.Check.RequestBody = true

		default:
			return d.Errf("unrecognized validate option: '%s'", token)
		}
	}

	return nil
}

//err := parseValidate(oapi, d)
