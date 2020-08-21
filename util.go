package openapi

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func (oapi OpenAPI) log(msg string) {
	defer oapi.logger.Sync()

	sugar := oapi.logger.Sugar()
	sugar.Infof(msg)
}

func parseValidateDirective(oapi *OpenAPI, d *caddyfile.Dispenser) error {

	args := d.RemainingArgs()
	if len(args) <= 0 {
		return d.ArgErr()
	}

	for _, token := range args {

		switch token {
		case VALUE_REQUEST_PARAMS:
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
