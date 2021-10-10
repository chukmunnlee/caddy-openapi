package openapi

import (
	"fmt"
	"strings"

	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func (oapi OpenAPI) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {

	url := req.URL
	if oapi.ValidateServers {
		url.Host = req.Host
		if nil == req.TLS {
			url.Scheme = "http"
		} else {
			url.Scheme = "https"
		}
	}

	replacer := req.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	replacer.Set(OPENAPI_ERROR, "")
	replacer.Set(OPENAPI_STATUS_CODE, "")

	route, pathParams, err := oapi.router.FindRoute(req)

	if nil != err {
		replacer.Set(OPENAPI_ERROR, err.Error())
		replacer.Set(OPENAPI_STATUS_CODE, 404)
		if oapi.LogError {
			oapi.err(fmt.Sprintf("%s %s %s: %s", getIP(req), req.Method, req.RequestURI, err))
		}
		if !oapi.FallThrough {
			return err
		}
	}

	// don't check if we have a 404 on the route
	if (nil == err) && (nil != oapi.Check) {
		if oapi.Check.RequestParams {
			validateReqInput := &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					ExcludeRequestBody: !oapi.Check.RequestBody,
				},
			}
			err = openapi3filter.ValidateRequest(req.Context(), validateReqInput)
			if nil != err {
				reqErr := err.(*openapi3filter.RequestError)
				replacer.Set(OPENAPI_ERROR, reqErr.Error())
				//replacer.Set(OPENAPI_STATUS_CODE, reqErr.HTTPStatus())
				replacer.Set(OPENAPI_STATUS_CODE, 400)
				if oapi.LogError {
					oapi.err(fmt.Sprintf(">> %s %s %s: %s", getIP(req), req.Method, req.RequestURI, err))
				}
				if !oapi.FallThrough {
					return err
				}
			}
		}
	}

	if query, exists := resolvePolicy(route, req.Method); exists {
		result, err := evalPolicy(query, oapi.policy, req, pathParams)
		if nil != err {
			replacer.Set(OPENAPI_ERROR, err.Error())
			replacer.Set(OPENAPI_STATUS_CODE, 403)
			if oapi.LogError {
				oapi.err(err.Error())
			}
			return nil
		}

		if !result {
			err = fmt.Errorf("Denied: %s", query)
			replacer.Set(OPENAPI_ERROR, err.Error())
			replacer.Set(OPENAPI_STATUS_CODE, 403)
			if oapi.LogError {
				oapi.err(err.Error())
			}
			return err
		}
	}

	wrapper := &WrapperResponseWriter{ResponseWriter: w}
	if err := next.ServeHTTP(wrapper, req); nil != err {
		return err
	}

	if nil != oapi.contentMap {
		contentType := w.Header().Get("Content-Type")
		if "" == contentType {
			return nil
		}
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		_, ok := oapi.contentMap[contentType]
		if !ok {
			return nil
		}

		validateReqInput := &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				ExcludeRequestBody:    true,
				ExcludeResponseBody:   false,
				IncludeResponseStatus: true,
			},
		}

		if (nil != wrapper.Buffer) && (len(wrapper.Buffer) > 0) {
			validateRespInput := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: validateReqInput,
				Status:                 wrapper.StatusCode,
				Header:                 http.Header{"Content-Type": oapi.Check.ResponseBody},
			}
			validateRespInput.SetBodyBytes(wrapper.Buffer)
			if err := openapi3filter.ValidateResponse(req.Context(), validateRespInput); nil != err {
				respErr := err.(*openapi3filter.ResponseError)
				oapi.err(fmt.Sprintf("<< %s %s %s: %s", getIP(req), req.Method, req.RequestURI, respErr.Error()))
			}
		}
	}
	return nil
}
