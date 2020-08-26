package openapi

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

const (
	OPENAPI_ERROR       = "openapi.error"
	OPENAPI_STATUS_CODE = "openapi.status_code"
	TOKEN_OPENAPI       = "openapi"
	TOKEN_SPEC          = "spec"
	TOKEN_FALL_THROUGH  = "fall_through"
	TOKEN_LOG_ERROR     = "log_error"
	TOKEN_VALIDATE      = "validate"
	VALUE_REQ_PARAMS    = "req_params"
	VALUE_REQ_BODY      = "req_body"
)

type OpenAPI struct {
	Spec            string   `json:"spec"`
	FallThrough     bool     `json:"fall_through,omitempty"`
	LogError        bool     `json:"log_error,omitempty"`
	RequestParams   bool     `json:"req_params,omitempty"`
	RequestBody     bool     `json:"req_body,omitempty"`
	RequestBodyType []string `json:"req_body_type,omitempty"`

	swagger *openapi3.Swagger
	router  *openapi3filter.Router

	logger *zap.Logger
}

var (
	_ caddy.Provisioner           = (*OpenAPI)(nil)
	_ caddy.Validator             = (*OpenAPI)(nil)
	_ caddyfile.Unmarshaler       = (*OpenAPI)(nil)
	_ caddyhttp.MiddlewareHandler = (*OpenAPI)(nil)
)

func init() {
	caddy.RegisterModule(OpenAPI{})
	httpcaddyfile.RegisterHandlerDirective("openapi", parseCaddyFile)
}

func (oapi OpenAPI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.openapi",
		New: func() caddy.Module { return new(OpenAPI) },
	}
}

func (oapi *OpenAPI) Provision(ctx caddy.Context) error {

	var swagger *openapi3.Swagger
	var err error
	var err2 error

	oapi.logger = ctx.Logger(oapi)
	defer oapi.logger.Sync()

	oapi.log(fmt.Sprintf("Using OpenAPI spec: %s", oapi.Spec))

	if strings.HasPrefix("http", oapi.Spec) {
		var u *url.URL
		if u, err = url.Parse(oapi.Spec); nil != err {
			return err
		}
		if swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromURI(u); nil != err {
			return err
		}
	} else if _, err = os.Stat(oapi.Spec); !(nil == err || os.IsExist(err)) {
		return err

	} else if swagger, err2 = openapi3.NewSwaggerLoader().LoadSwaggerFromFile(oapi.Spec); nil != err2 {
		return err2
	}

	router := openapi3filter.NewRouter()
	router.AddSwagger(swagger)

	oapi.swagger = swagger
	oapi.router = router

	return nil
}

func (oapi OpenAPI) Validate() error {
	return nil
}

func (oapi *OpenAPI) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {

	oapi.Spec = ""
	oapi.FallThrough = false
	oapi.LogError = false
	oapi.RequestParams = false
	oapi.RequestBody = false

	// Skip the openapi directive
	d.Next()
	args := d.RemainingArgs()
	if 1 == len(args) {
		d.NextArg()
		oapi.Spec = d.Val()
	}

	for nest := d.Nesting(); d.NextBlock(nest); {
		token := d.Val()
		switch token {
		case TOKEN_SPEC:
			if !d.NextArg() {
				return d.Err("missing OpenAPI spec file")
			} else {
				oapi.Spec = d.Val()
			}
			if d.NextArg() {
				return d.ArgErr()
			}

		case TOKEN_FALL_THROUGH:
			if d.NextArg() {
				return d.ArgErr()
			}
			oapi.FallThrough = true

		case TOKEN_LOG_ERROR:
			if d.NextArg() {
				return d.ArgErr()
			}
			oapi.LogError = true

		case TOKEN_VALIDATE:
			err := parseValidateDirective(oapi, d)
			if nil != err {
				return err
			}

		default:
			return d.Errf("unrecognized subdirective: '%s'", token)
		}
	}

	if "" == oapi.Spec {
		return d.Err("missing OpenAPI spec file")
	}
	return nil
}

func parseCaddyFile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var oapi OpenAPI
	err := oapi.UnmarshalCaddyfile(h.Dispenser)
	return oapi, err
}
