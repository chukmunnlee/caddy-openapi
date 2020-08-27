package openapi

import (
	"fmt"

	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func (oapi OpenAPI) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {

	url := req.URL
	url.Host = req.Host
	if nil == req.TLS {
		url.Scheme = "http"
	} else {
		url.Scheme = "https"
	}

	replacer := req.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	replacer.Set(OPENAPI_ERROR, "")
	replacer.Set(OPENAPI_STATUS_CODE, "")

	route, pathParams, err := oapi.router.FindRoute(req.Method, url)
	if nil != err {
		replacer.Set(OPENAPI_ERROR, err.Error())
		replacer.Set(OPENAPI_STATUS_CODE, 404)
		if oapi.LogError {
			oapi.log(fmt.Sprintf("%s %s %s: %s", getIP(req), req.Method, req.RequestURI, err))
		}
		if !oapi.FallThrough {
			return err
		}
	}

	// don't check if we have a 404 on the route
	if (nil == err) && (nil != oapi.Check) && oapi.Check.RequestParams {
		validateParams := &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				ExcludeRequestBody: !oapi.Check.RequestBody,
			},
		}
		err = openapi3filter.ValidateRequest(req.Context(), validateParams)
		if nil != err {
			reqErr := err.(*openapi3filter.RequestError)
			replacer.Set(OPENAPI_ERROR, reqErr.Error())
			replacer.Set(OPENAPI_STATUS_CODE, reqErr.HTTPStatus())
			if oapi.LogError {
				oapi.log(fmt.Sprintf("%s %s %s: %s", getIP(req), req.Method, req.RequestURI, err))
			}
			if !oapi.FallThrough {
				return err
			}
		}
	}

	return next.ServeHTTP(w, req)
}

func (oapi OpenAPI) log(msg string) {
	defer oapi.logger.Sync()

	sugar := oapi.logger.Sugar()
	sugar.Infof(msg)
}
