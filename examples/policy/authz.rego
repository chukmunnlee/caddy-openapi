package authz

default allow = false

allow {
	lower(input.method) = "get"
	array.slice(input.path, 0, 2) = [ "api", "customer" ]
	to_number(input.pathParams.custId) >= 100
}
