# caddy-openapi

This middleware validates HTTP request and response against a OpenAPI V3 Specification file

## Installation

Build caddy with caddy-openapi, run `make`. This will build for Linux, Windows and OSX.

You can also build with `xcaddy`
```
xcaddy build \
    --with github.com/chukmunnlee/caddy-openapi
```

Tested with `go version go1.17.1 linux/amd64`.

## Usage

### Caddyfile

Load `examples/customer/customer.yaml` file with defaults

```
:8080 {
  route /api {
    openapi ./examples/customer/customer.yaml
  }
}
```

One with all the options

```
:8080 {
  route /api {
    openapi {
      spec ./examples/customer/customer.yaml
      fall_through
      log_error
    }
  }
}
```

Reports any errors as a `{openapi.error}` [placeholder](https://caddyserver.com/docs/caddyfile/concepts#placeholders) which can be used in other [directives](https://caddyserver.com/docs/caddyfile/directives) like [`respond`](https://caddyserver.com/docs/caddyfile/directives/respond)

| Fields                   | Description |
|--------------------------|-------------|
| `spec <oas_file>`        | The OpenAPI3 YAML file. This attribute is a mandatory |
| `policy_bundle <bundle>` | [OPA](https://www.openpolicyagent.org/) policy bundle created with `opa build` |
| `fall_through`           | Toggles fall through when the request does do match the provided OpenAPI spec. Default is `false` |
| `validate_servers`       | Enable server validation. Accepts `true`, `false` or just the directive which enables validation. Default is `true`. |
| `log_error`              | Toggles error logging. Default is `false` |
| `check`                  | Enable validation of the request parameters; include one or more of the following directives in the body:`req_params`, `req_body` and `resp_body`. `resp_body` only validates `application/json` payload. Note that validating the request body will implicitly set `req_params` |

Errors are reported in the following 2 [placeholder](https://caddyserver.com/docs/caddyfile/concepts#placeholders). You can use them in other [directives](https://caddyserver.com/docs/caddyfile/directives) like [`respond`](https://caddyserver.com/docs/caddyfile/directives/respond)

| Placeholders          | Description |
|-----------------------|-------------|
| `openapi.error`       | Description of the error |
| `openapi.status_code` | Suggested status code |


Reports any errors as a `{openapi.error}` 

## Example

The following example validates all request, including query string as well as payloads, to `localhost:8080/api` 
against the `./examples/customer/customer.yaml` file.  Any non compliant request will be logged to Caddy's console. 
Respond to the client with the error `{openapi.error}`.

```
:8080 {

  @api {
    path /api/*
  }

  reverse_proxy @api {
    to localhost:3000
  }

  route @api {
    openapi {
      spec ./examples/customer/customer.yaml 
      policy_bundle ./examples/policy/bundle.tar.gz
      check {
        req_body 
        resp_body 
      }
      validate_servers
      log_error 
    }
  }

  handle_errors {
    respond @api "Resource: {http.request.orig_uri}. Error: {openapi.error}" {openapi.status_code}  {
      close
    }
  }
}
```

Try out the `customer.yaml` API by running the accompanying node application.

## Using OpenPolicyAgent

You can enforce policies on routes by adding the `x-policy` field to either the [OpenAPI3 document](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#schema) level, or the [path item](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#pathItemObject) level or or the (operation)[https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#operationObject) level. 

If a `x-policy` field is added at the
- *OpenAPI3 document* then the policy will be applied to all path
- *Path item* then the policy will be applied to all methods specified for that path eg `POST`, `GET` to `/api/v1/customer`
- *Operation* then the policy will only be applied to that operation eg. `GET/api/v1/customer`
`x-policy` attribute nested deeper into the 

The 'deeper' a `x-policy` field, the higher its precedence. 

Assume the following OPA policy file
```
package authz

default allow = false

allow {
  lower(input.method) = "get"
  array.slice(input.path, 0, 2) = [ "api", "customer" ]
  to_number(input.pathParams.custId) >= 100
}
```
has been bundled as `bundle.tar.gz`.

The following OpenAPI3 fragment show how you can evaluate `authz.allow` on all `GET /api/customer/`
```
paths:
  /api/customer/{custId}:
    get:
      description: Get customer
      operationId: getCustomer
      x-policy: authz.allow
      parameters:
      - name: custId
        in: path
        required: true
        schema:
           type: number
```

The HTTP request are converted into `input` according to the following table

| Fields                   | Description |
|--------------------------|-------------|
| `input.scheme`           | HTTP or HTTPS |
| `input.host`             | Host and port number |
| `input.method`           | HTTP method  |
| `input.path`             | Array of path elements eg. `/api/customer/123` is converted to `[ 'api', 'customer', '123 ]` |
| `input.remoteAddr`       | Host and port number of the client |
| `input.queryString`      | If a query string is present, the query string will be destructed into a map under `queryString` root. Example `?offset=10&limit=10` will be converted to the following keys: `input.queryString.offset` and `input.queryString.limit`. Query parameters with multiple value will have an array as its value. `queryString` will not be present if the request do not contain any query params |
| `input.pathParams`       | Like query string but a map of matched path parameters from the OpenAPI3 spec where parameter type is `in: path`. See above example |
| `input.headers`          | Map of all the request headers |
| `input.body`             | Access to the request's body. Only supports `application/json` content type. **Not implemented yet** |

Assume all values are string
