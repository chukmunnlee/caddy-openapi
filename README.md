# caddy-openapi

This middleware validates HTTP request and response against a OpenAPI V3 Specification file

## Installation

```
xcaddy build v2.1.1 \
    --with github.com/chukmunnlee/caddy-openapi
```

## Usage

### Caddyfile

Load `samples/hello.yaml` file with defaults

```
:8080 {
  route /api {
    openapi ./samples/hello.yaml
  }
}
```

One with all the options

```
:8080 {
  route /api {
    openapi {
      spec ./samples/hello.yaml
      fall_through
      log_error
    }
  }
}
```

Reports any errors as a `{openapi.error}` [placeholder](https://caddyserver.com/docs/caddyfile/concepts#placeholders) which can be used in other [directives](https://caddyserver.com/docs/caddyfile/directives) like [`respond`](https://caddyserver.com/docs/caddyfile/directives/respond)

| Fields            | Description |
|-------------------|-------------|
| `spec <oas_file>` | The OpenAPI file to use. Overrides the file used with the `openapi` directive |
| `fall_through`    | Toggles fall through when the request does do match the provided OpenAPI spec. Default is `false` |
| `log_error`       | Toggles error logging. Default is `false` |
| `validate`        | Enable validation of the request parameters `req_params` and/or`req_body`. Note that validating the request body will implicitly set `req_params` |

Errors are reported in the following 2 [placeholder](https://caddyserver.com/docs/caddyfile/concepts#placeholders). You can use them in other [directives](https://caddyserver.com/docs/caddyfile/directives) like [`respond`](https://caddyserver.com/docs/caddyfile/directives/respond)

| Placeholders          | Description |
|-----------------------|-------------|
| `openapi.error`       | Description of the error |
| `openapi.status_code` | Suggested status code |


Reports any errors as a `{openapi.error}` 

## Example

The following example validates all request, including query string as well as payloads, to `localhost:8080/api` 
against `./samples/hello.yaml` file.  Any non compliant request will be logged to Caddy's console. 
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
      spec ./samples/hello.yaml 
		validate req_body 
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

