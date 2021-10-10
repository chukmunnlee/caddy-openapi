package openapi

import (
	"context"
	"fmt"
	"strings"

	"net/http"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/open-policy-agent/opa/rego"
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
			return d.Errf("Unrecognized validate option: '%s'", token)
		}
	}

	return nil
}

func resolvePolicy(r *routers.Route, method string) (string, bool) {

	// global
	policy := hasXPolicy(r.Spec.ExtensionProps, "")

	var oper *openapi3.Operation

	// path
	policy = hasXPolicy(r.PathItem.ExtensionProps, policy)

	switch strings.ToLower(method) {
	case "get":
		oper = r.PathItem.Get
	case "delete":
		oper = r.PathItem.Delete
	case "head":
		oper = r.PathItem.Head
	case "options":
		oper = r.PathItem.Options
	case "patch":
		oper = r.PathItem.Patch
	case "post":
		oper = r.PathItem.Post
	case "put":
		oper = r.PathItem.Put
	case "trace":
		oper = r.PathItem.Trace
	default:
		return policy, "" != policy
	}

	// method
	policy = hasXPolicy(oper.ExtensionProps, policy)

	return policy, "" != policy
}

func hasXPolicy(p openapi3.ExtensionProps, d string) string {
	v := p.Extensions[X_POLICY]
	if nil == v {
		return d
	}
	q := fmt.Sprintf("%s", v)
	return strings.ReplaceAll(q, "\"", "")
}

func evalPolicy(query string, policy func(*rego.Rego), req *http.Request, pathParams map[string]string) (bool, error) {

	input := mkInput(req, pathParams)

	ctx := context.TODO()

	eval, err := rego.New(rego.Query(fmt.Sprintf("data.%s", query)), policy).PrepareForEval(ctx)
	if nil != err {
		return false, err
	}

	result, err := eval.Eval(ctx, rego.EvalInput(input))
	if nil != err {
		return false, err
	} else if 0 == len(result) {
		// undefined did not have a default
		return false, nil
	}

	return result.Allowed(), nil
}

func mkInput(req *http.Request, pathParams map[string]string) map[string]interface{} {

	url := req.URL

	input := make(map[string]interface{})

	input["scheme"] = url.Scheme
	input["host"] = req.Host
	input["method"] = req.Method
	input["path"] = strings.Split(url.Path[1:], "/")

	input["remoteAddr"] = req.RemoteAddr

	if len(url.Query()) > 0 {
		q := make(map[string]interface{})
		for k, v := range req.URL.Query() {
			if len(v) == 1 {
				q[k] = v[0]
			} else {
				q[k] = v
			}
		}
		input["queryString"] = q
	}

	if len(req.Header) > 0 {
		h := make(map[string]interface{})
		for k, v := range req.Header {
			if len(v) == 1 {
				h[k] = v[0]
			} else {
				h[k] = v
			}
		}
		input["headers"] = h
	}

	if len(pathParams) > 0 {
		input["pathParams"] = pathParams
	}

	return input
}

func (oapi OpenAPI) log(msg string) {
	defer oapi.logger.Sync()

	sugar := oapi.logger.Sugar()
	sugar.Infof(msg)
}

func (oapi OpenAPI) err(msg string) {
	defer oapi.logger.Sync()
	sugar := oapi.logger.Sugar()
	sugar.Errorf(msg)
}
