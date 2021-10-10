package openapi

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/open-policy-agent/opa/rego"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const (
	MODULE_ID              = "http.handlers.openapi"
	X_POLICY               = "x-policy"
	OPENAPI_ERROR          = "openapi.error"
	OPENAPI_STATUS_CODE    = "openapi.status_code"
	TOKEN_OPENAPI          = "openapi"
	TOKEN_POLICY_BUNDLE    = "policy_bundle"
	TOKEN_SPEC             = "spec"
	TOKEN_FALL_THROUGH     = "fall_through"
	TOKEN_LOG_ERROR        = "log_error"
	TOKEN_VALIDATE_SERVERS = "validate_servers"
	TOKEN_CHECK            = "check"
	VALUE_REQ_PARAMS       = "req_params"
	VALUE_REQ_BODY         = "req_body"
	VALUE_RESP_BODY        = "resp_body"
)

// This middleware validates request against an OpenAPI V3 specification. No conforming request can be rejected
type OpenAPI struct {
	// The location of the OASv3 file
	Spec string `json:"spec"`

	PolicyBundle string `json:"policy_bundle"`

	// Should the request proceed if it fails validation. Default is `false`
	FallThrough bool `json:"fall_through,omitempty"`

	// Should the non compliant request be logged? Default is `false`
	LogError bool `json:"log_error,omitempty"`

	// Enable request and response validation
	Check *CheckOptions `json:"check,omitempty"`

	// Enable server validation
	ValidateServers bool `json:"valid_servers,omitempty"`

	oas    *openapi3.T
	router routers.Router

	logger *zap.Logger

	contentMap map[string]string

	policy func(*rego.Rego)
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
	httpcaddyfile.RegisterHandlerDirective(TOKEN_OPENAPI, parseCaddyFile)
}

func (oapi OpenAPI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  MODULE_ID,
		New: func() caddy.Module { return new(OpenAPI) },
	}
}

func (oapi *OpenAPI) Provision(ctx caddy.Context) error {

	var oas *openapi3.T
	var err error

	oapi.logger = ctx.Logger(oapi)
	defer oapi.logger.Sync()

	oapi.log(fmt.Sprintf("Using OpenAPI spec: %s", oapi.Spec))

	if strings.HasPrefix("http", oapi.Spec) {
		var u *url.URL
		if u, err = url.Parse(oapi.Spec); nil != err {
			return err
		}
		if oas, err = openapi3.NewLoader().LoadFromURI(u); nil != err {
			return err
		}
	} else if _, err = os.Stat(oapi.Spec); !(nil == err || os.IsExist(err)) {
		return err

	} else if oas, err = openapi3.NewLoader().LoadFromFile(oapi.Spec); nil != err {
		return err
	}

	if oapi.ValidateServers {
		oapi.log("List of servers")
		for _, s := range oas.Servers {
			oapi.log(fmt.Sprintf("- %s #%s", s.URL, s.Description))
		}
	} else {
		// clear all servers
		oapi.log("Disabling server validation")
		oas.Servers = make([]*openapi3.Server, 0)
	}

	router, err := gorillamux.NewRouter(oas)

	if nil != err {
		return err
	}

	oapi.oas = oas
	oapi.router = router

	if (nil != oapi.Check) && (nil != oapi.Check.ResponseBody) {
		oapi.contentMap = make(map[string]string)
		for _, content := range oapi.Check.ResponseBody {
			oapi.contentMap[content] = ""
		}
	}

	if len(oapi.PolicyBundle) > 0 {
		oapi.log(fmt.Sprintf("Loaded policy bundle: %s", oapi.PolicyBundle))
		oapi.policy = rego.LoadBundle(oapi.PolicyBundle)
	}

	return nil
}

func (oapi OpenAPI) Validate() error {
	return nil
}

func (oapi *OpenAPI) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {

	oapi.Spec = ""
	oapi.PolicyBundle = ""
	oapi.FallThrough = false
	oapi.LogError = false
	oapi.ValidateServers = true
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
				return d.Err("Missing OpenAPI spec file")
			} else {
				oapi.Spec = d.Val()
			}
			if d.NextArg() {
				return d.ArgErr()
			}

		case TOKEN_POLICY_BUNDLE:
			if !d.NextArg() {
				return d.Err("Missing policy bundle")
			} else {
				oapi.PolicyBundle = d.Val()
			}
			if d.NextArg() {
				return d.ArgErr()
			}

		case TOKEN_VALIDATE_SERVERS:
			if d.NextArg() {
				b, err := strconv.ParseBool(d.Val())
				if nil == err {
					oapi.ValidateServers = b
				}
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
