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
	OPENAPI_ERROR = "openapi.error"
)

type OpenAPI struct {
	Spec string `json:"spec"`

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

func (oas *OpenAPI) Provision(ctx caddy.Context) error {
	fmt.Println("in Provision")

	var swagger *openapi3.Swagger
	var err error
	var abc error

	oas.logger = ctx.Logger(oas)
	defer oas.logger.Sync()

	oas.log(fmt.Sprintf("Using OpenAPI spec: %s", oas.Spec))

	if strings.HasPrefix("http", oas.Spec) {
		var u *url.URL
		if u, err = url.Parse(oas.Spec); nil != err {
			return err
		}
		if swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromURI(u); nil != err {
			return err
		}
	} else if _, err = os.Stat(oas.Spec); !(nil == err || os.IsExist(err)) {
		return err

	} else if swagger, abc = openapi3.NewSwaggerLoader().LoadSwaggerFromFile(oas.Spec); nil != abc {
		return abc
	}

	router := openapi3filter.NewRouter()
	router.AddSwagger(swagger)

	oas.swagger = swagger
	oas.router = router

	return nil
}

func (oas OpenAPI) Validate() error {
	return nil
}

func (oas *OpenAPI) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()

		switch len(args) {
		case 0:
			return d.Err("missing openapi specification file")
		case 1:
			oas.Spec = args[0]
		default:
			return d.Err("unexpected number of arguments")
		}
	}

	return nil
}

func parseCaddyFile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var oapi OpenAPI
	err := oapi.UnmarshalCaddyfile(h.Dispenser)
	return oapi, err
}
