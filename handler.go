package openapi

import (
	"fmt"

	"net/http"

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

	//route, pathParams, err := oapi.router.FindRoute(req.Method, url)
	_, _, err := oapi.router.FindRoute(req.Method, url)
	if nil != err {
		replacer.Set(OPENAPI_ERROR, err.Error())
		oapi.log(fmt.Sprintf("%s: %s:%s", err, req.Method, req.RequestURI))
		return err
	} else {
		replacer.Delete(OPENAPI_ERROR)
	}

	return next.ServeHTTP(w, req)
}
