package authz 

req_0 = {
	"method": "GET",
	"path": [ "api", "customer", 100 ],
	"pathParams": {
		"custId": 100
	}
}

req_1 = {
	"method": "GET",
	"path": [ "api", "customer", 90 ],
	"pathParams": {
		"custId": 90
	}
}

test_allow_get_customer_100  {
	allow with input as req_0
}

#test_not_allow_get_customer_90  {
#	not allow with input as req_1
#}

