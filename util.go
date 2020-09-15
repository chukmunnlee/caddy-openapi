package openapi

import (
	"strings"

	"net/http"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

type WrapperResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Buffer     []byte
}

func (w *WrapperResponseWriter) WriteHeader(sc int) {
	w.ResponseWriter.WriteHeader(sc)
	w.StatusCode = sc
}

func (w *WrapperResponseWriter) Write(buff []byte) (int, error) {
	w.Buffer = append(w.Buffer[:], buff[:]...)
	return w.ResponseWriter.Write(buff)
}

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

	oapi.Check = &CheckOptions{RequestBody: false, RequestParams: false, ResponseBody: nil}

	for nest := d.Nesting(); d.NextBlock(nest); {
		token := d.Val()
		switch token {
		case VALUE_REQ_PARAMS:
			oapi.Check.RequestParams = true

		case VALUE_REQ_BODY:
			oapi.Check.RequestParams = true
			oapi.Check.RequestBody = true

		case VALUE_RESP_BODY:
			args := d.RemainingArgs()
			oapi.Check.ResponseBody = make([]string, len(args))
			if len(args) <= 0 {
				oapi.Check.ResponseBody = append(oapi.Check.ResponseBody, "application/json")
			} else {
				for i, content := range args {
					oapi.Check.ResponseBody[i] = strings.ToLower(strings.TrimSpace(content))
				}
			}

		default:
			return d.Errf("unrecognized validate option: '%s'", token)
		}
	}

	return nil
}

func (oapi OpenAPI) log(msg string) {
	defer oapi.logger.Sync()

	sugar := oapi.logger.Sugar()
	sugar.Infof(msg)
}
