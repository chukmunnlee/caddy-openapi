# caddy-openapi

This middleware validates HTTP request and response against a OpenAPI V3 Specification file

## Installation

```
xcaddy build v2.1.1 \
    --with github.com/chukmunnlee/caddy-openapi
```

## Usage

Reports any errors as a `{openapi.error}` [placeholder](https://caddyserver.com/docs/caddyfile/concepts#placeholders) which can be used in other [directives](https://caddyserver.com/docs/caddyfile/directives) like [`respond`](https://caddyserver.com/docs/caddyfile/directives/respond)

```
openapi oas3_file.yaml
```

where **`oas3_file.yaml`** is a OASv3 file used for validation.

#### Example

```
:8080 {

	@api {
		path /api/*
	}

	reverse_proxy @api {
		to localhost:3000
	}

	route @api {
		openapi ./samples/hello.yaml 
	}

	handle_errors {
		respond @api "Resource: {http.request.orig_uri}. Error: {openapi.error}" 400  {
			close
		}
	}
}
```

## Features

* Validate request's method and resource
