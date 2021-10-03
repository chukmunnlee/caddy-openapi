const morgan = require('morgan')
const express = require('express')

const { findCustomer, addCustomer, getCustomers } =  require('./customerdb')

const PORT = parseInt(process.env.PORT) || 3000

const app = express()

app.use(morgan('common'))

app.get('/api/customer/:custId', 
	(req, resp) => {
		const custId = req.params.custId
		const rec = findCustomer(custId)
		if (rec) 
			return resp.status(200).type('application/json').json(rec)

		resp.status(404).type('application/json')
			.json({ message: `Cannot find customer ${custId}` })
	}
)

app.get('/api/customers',
	(req, resp) => {
		const offset = parseInt(req.query.offset) || 0
		const limit = parseInt(req.query.limit) || 3
		resp.status(200).type('application/json')
			.json(getCustomers(offset, offset + limit))
	}
)

app.post('/api/customer', 
	express.urlencoded({ extended: true }), express.json(),
	(req, resp) => {
		const rec = findCustomer(req.body.customerId)
		if (rec)
			return resp.status(400).type('application/json')
				.json({ message: `Cannot insert customer`})

		addCustomer(req.body)
		resp.status(201).type('application/json')
			.json({
				message: `Added customer ${req.body.customerId}`,
				time: (new Date()).toLocaleString()
			})
	}
)

app.listen(PORT, () => {
	console.info(`Application started on port ${PORT} at ${new Date()}`)
})
