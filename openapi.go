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
	TOKEN_CHECK         = "check"
	VALUE_REQ_PARAMS    = "req_params"
	VALUE_REQ_BODY      = "req_body"
	VALUE_RESP_BODY     = "resp_body"
)

// This middleware validates request against an OpenAPI V3 specification. No conforming request can be rejected
type OpenAPI struct {
	// The location of the OASv3 file
	Spec string `json:"spec"`

	// Should the request proceed if it fails validation. Default is `false`
	FallThrough bool `json:"fall_through,omitempty"`

	// Should the non compliant request be logged? Default is `false`
	LogError bool `json:"log_error,omitempty"`

	// Enable request and response validation
	Check *CheckOptions `json:"check,omitempty"`

	swagger *openapi3.Swagger
	router  *openapi3filter.Router

	logger *zap.Logger

	contentMap map[string]string
}

type CheckOptions struct {
	// Enable request query validation. Default is `false`
	RequestParams bool `json:"req_params,omitempty"`

	// Enable request payload validation. Default is `false`
	RequestBody bool `json:"req_body,omitempty"`

	// Enable response body validation with an optional list of
	// `Content-Type` to examine. Default `application/json`. If you set
	// your content type, the default will be removed
	ResponseBody []string `json:"resp_body,omitempty"`
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

	if (nil != oapi.Check) && (nil != oapi.Check.ResponseBody) {
		oapi.contentMap = make(map[string]string)
		for _, content := range oapi.Check.ResponseBody {
			oapi.contentMap[content] = ""
		}
	}

	return nil
}

func (oapi OpenAPI) Validate() error {
	return nil
}

func (oapi *OpenAPI) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {

	oapi.Spec = ""
	oapi.FallThrough = false
	oapi.LogError = false
	oapi.Check = nil

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

		case TOKEN_CHECK:
			err := parseCheckDirective(oapi, d)
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
