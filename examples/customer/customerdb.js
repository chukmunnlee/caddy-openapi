const records = [
	{
		customerId: 100,
		name: 'fred',
		email: 'fred@gmail.com'
	},
	{
		customerId: 101,
		name: 'barney',
		email: 'barney@gmail.com'
	},
	{
		customerId: 102,
		name: 'wilma',
		email: 'wilma@gmail.com'
	},
	{
		customerId: 103,
		name: 'betty',
		email: 'betty@gmail.com'
	},
]

const findCustomer = (custId) => {
	return records.find(v => v.customerId == custId) || false
}

const addCustomer = (rec) => {
	records.push(rec)
}

const getCustomers = (offset = 0, limit = 3) => {
	return records.slice(offset, limit)
}

module.exports = { findCustomer, addCustomer, getCustomers }
