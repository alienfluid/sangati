# Sangati

*Sangati*, the Hindi term for *consistency*, is a command line tool to makes it easier to check for logical inconsistencies or business logic errors in your data. It does so by automatically executing a series of SQL statements and comparing the output to either values that are statically defined in a configuration file or the output of another SQL query. 

While ideally these checks would be implemented as constraints on the respective tables in an RDBMS or as descriptive relationships in the ORM of choice, frequently it is not practical to do so. For example, the code might have been ported from a legacy system that did not use an ORM or the data might be sourced from an ETL process that does not respect the constraints of the source database. In such cases, Sangati provides a low-cost way for checking for obvious data integrity issues at some regular frequency (e.g. as part of the release process or nightly).

*Sangati* can act as a light-weight quality assurance tool in your overall Master Data Management (MDM) strategy. For example, it can be used to assure that no duplicates ever exist in the master data or that measures are standardized by checking for the presence of extreme values. Used this way, it can help with data governance and ensure that the organization keeps its master data clean. Being a command line tool that outputs to `STDOUT` and `STDERR`, it can be easily integrated within the existing MDM/data governance toolchain. 

Currently, *Sangati* only supports connections to PostgreSQL and its variants. Support for other database systems will be added in the future.

## Examples of logical constraints that can be tested for

* Does every user row have an associated company?
* Do the number of records added yesterday match in the production and data warehouse databases?
* Were there any new records added to the sales table in the last day?
* Are the number of events generated in the last hour below a certain expected threshold?

## To download and compile

```bash
go get -v github.com/alienfluid/sangati
go install github.com/alienfluid/sangati
```

## Usage

```bash
./sangati -c="path/to/configuration/file.json"
```

## Sample configuration file

```json
{
    "Databases" : [
                    {
                        "Host": "host1.db.mydomain.com",
                        "Port": 5432,
                        "DbName": "public",
                        "Index": 1
                    },
                    {   "Host": "backup.db.mydomain.com",
                        "Port": 5432,
                        "DbName": "default",
                        "Index": 2
                    }
                ],
    "Tests": [
                {
                    "Name": "Compare output of one query against one value",
                    "Types": ["int"],
                    "Queries": [
                                {
                                    "DbIndex": 1, 
                                    "Query": "SELECT COUNT(1) FROM users"
                                }
                            ],
                    "Values": ["0"],
                    "Operator": "gt"
                },
                {
                    "Name": "Compare output of one query against multiple values",
                    "Types": ["int", "string"],
                    "Queries": [
                                {
                                    "DbIndex": 1, 
                                    "Query": "SELECT name, email FROM users WHERE id = 4"
                                }
                            ],
                    "Values": ["Farhan Ahmed", "some@email.com"],
                    "Operator": "eq"
                },
                {
                    "Name": "Compare output of one query against output of another",
                    "Types": ["date", "int"],
                    "Queries": [
                                {
                                    "DbIndex": 1,
                                    "Query": "SELECT create_date, COUNT(1) FROM companies GROUP BY 1 ORDER BY 1"
                                },
                                {
                                    "DbIndex": 2,
                                    "Query": "SELECT create_date, COUNT(1) FROM companies GROUP BY 1 ORDER BY 1"
                                }
                            ],
                    "Values": []
                }               
            ]
}

```

## Single vs. Multi-query tests

Sangati has two modes of operation and they are controlled by the structure of the individual tests in the configuration file.

### Single query tests

Single query tests compare the output of a query to the static value(s) specified in the configuration file. The output of the query must be a single row of data, however multiple columns are supported. The data must be of one of the following types -

* `string` (VARCHAR)
* `int` (INT32)
* `date` (DATE)
* `timestamp` (TIMESTAMP) (of the format `2006-02-01 15:04:05.000000`)

The following operators are available to compare the output of the query to the value(s) specified in the configuration file.

```csv
lt      Less than
lte     Less than or equal to
gt      Greater than
gte     Greater than or equal to
eq      Equal to
```

Note that the operator applies to ALL the columns.

#### Example

```json
    {
        "Name": "Compare output of one query against one value",
        "Types": ["int"],
        "Queries": [
                    {
                        "DbIndex": 1, 
                        "Query": "SELECT COUNT(1) FROM users"
                    }
                ],
        "Values": ["0"],
        "Operator": "gt"
    }
```

This test compares the output of the query (which is an `integer`) to the static value `0` and fails the test if the returned value is _not_ greater than `0`.

### Multi-query tests

Multi-query tests compare the output of one SQL statement to the output of another. The queries can be executed against different databases. Multi-query tests allow support the comparison of multiple rows as as well multiple columns (matrix), however `equality` is the only supported logical operator in such tests. 

#### Example

```json
    {
        "Name": "Compare output of one query against output of another",
        "Types": ["date", "int"],
        "Queries": [
                    {
                        "DbIndex": 1,
                        "Query": "SELECT create_date, COUNT(1) FROM companies GROUP BY 1 ORDER BY 1"
                    },
                    {
                        "DbIndex": 2,
                        "Query": "SELECT create_date, COUNT(1) FROM companies GROUP BY 1 ORDER BY 1"
                    }
                ],
        "Values": []
    } 
```    

This compares the output of the first query to the output of the second query (which runs on a different database, see `DbIndex`) and fails the test is the outputs are not equal. Equality is the only supported predicate in multi-query tests.

## Frequently Asked Questions

* How do I specify the databases where the queries must be executed?

```json
    {
        "Host": "myhost.somedomain.com",
        "Port": 5432,
        "DbName": "main",
        "Index": 1
    }
```

The username and password must be specified as environment variables `DBUSER1` and `DBPASS1` respectively. If the environment variables are not set, Sangati assumes that the username and password are empty. 

When specifying multiple databases, the name of the environment variables to be set is the concatenation of `DBUSER` and `DBPASS` with the value of the `Index` field. For example, for a database with `Index` set to `2`, the environment variables `DBUSER2` and `DBPASS2` must be specified for Sangati to successfully connect to it.

* What types of data can be compared?

Currently, Sangati supports three data types --

```csv
string (VARCHAR)
int (INT32)
date (DATE)
timestamp (TIMESTAMP) (of the format 2006-02-01 15:04:05.000000)
```

* What databases do you currently support?

    Only PostgreSQL is currently supported. Support for other databases is coming soon.
