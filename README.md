# Sangati

*Sangati*, the Hindi term for *consistency*, is a command line tool to makes it easier to check for logical inconsistencies or business logic errors in your data. It does so by automatically executing a series of SQL statements (provided in a JSON-formatted configuration file) that each output a single _count_ as a result. This result is then compared to either static values or the output of another SQL statement using logical operators that are also specified in the configuration file. 

While ideally these checks would be implemented as constraints on the respective tables in an RDBMS or as descriptive relationships in the ORM of choice, frequently it is not practical to do so. For example, the code might have been ported from a legacy system that did not use an ORM or the data might be sourced from an ETL process that does not respect the constraints of the source database. In such cases, Sangati provides a low-cost way for checking for obvious data integrity issues at some regular frequency (e.g. as part of the release process or nightly).

Currently, *Sangati* only supports connections to PostgreSQL and its variants. Support for other database systems will be added in the future.

## Usage

```
./sangati -c="path/to/configuration/file.json"
```

## Sample configuration file

```
{
	"Host": "myhost.somedomain.com",
	"Port": 5432,
	"DbName": "main",
	"Tests": [
				{
					"Name": "At least one transaction in the last day",
					"Queries": ["SELECT COUNT(*) FROM transactions WHERE created_on > NOW() - INTERVAL '1 DAY'"],
					"Value": 0,
					"Operator": "gt"
				},
				{
					"Name": "Check for spikes in errors",
					"Queries": ["SELECT COUNT(*) FROM error_log WHERE created_on > NOW() - INTERVAL '1 DAY'"],
					"Value": 100000,
					"Operator": "lt"
				},
				{
					"Name": "Distinct users must be greater than on equal to distinct companies",
					"Queries": ["SELECT COUNT(DISTINCT id) FROM user",
								"SELECT COUNT(DISTINCT id) FROM company"],
					"Operator": "gte"
				}
			]
}
```

## Supported logical operators

```
lt 		Less than 
lte     Less than or equal to
gt      Greater than
gte     Greater than or equal to
eq      Equal to
```

## Examples

* Specifying the connection details for the database to connect to

```
	{
		"Host": "myhost.somedomain.com",
		"Port": 5432,
		"DbName": "main"
	}
```

The username and password must be specified as environment variables `DBUSER` and `DBPASS` respectively. If the environment variables are not set, Sangati assumes that the username and password are empty.

* Check for at least one new transaction since yesterday (_compare output to static value_)

```
	{
		"Name": "At least one transaction in the last day",
		"Queries": ["SELECT COUNT(*) FROM transactions WHERE created_on > NOW() - INTERVAL '1 DAY'"],
		"Value": 0,
		"Operator": "gt"
	}
```				

This test compares the output of the SQL statement to `0` and passes the test if the value is *greater than* `0` (as specified by the logical operator `gt`).

* Check high-level consistency in user and company tables (_compare output to another SQL statement_)

```
{
	"Name": "Distinct users must be greater than on equal to distinct companies",
	"Queries": ["SELECT COUNT(DISTINCT id) FROM user",
				"SELECT COUNT(DISTINCT id) FROM company"],
	"Operator": "gte"
}
```

This test compares the output of the first SQL statement to the output of the second SQL statement and passes the test if the first output is _greater than or equal to_ the second output (as specified by the logical operator `gte`).

## Frequently Asked Questions

* Where do I specify the username and password to connect to the database?

	These must be specified in environment variables `DBUSER` and `DBPASS` respectively.

* What databases do you currently support?

	Only PostgreSQL is currently supported. Support for other databases is coming soon.



