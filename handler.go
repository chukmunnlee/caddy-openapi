package openapi

import (
	"fmt"
	"strings"

	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func getIP(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if "" != ip {
		return strings.Split(ip, ",")[0]
	}
	return strings.Split(req.RemoteAddr, ":")[0]
}

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
		if oapi.LogError {
			oapi.log(fmt.Sprintf("%s %s %s: %s", getIP(req), req.Method, req.RequestURI, err))
		}
		if !oapi.FallThrough {
			return err
		}
	} else {
		replacer.Set(OPENAPI_ERROR, "")
	}

	return next.ServeHTTP(w, req)
}
