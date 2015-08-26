# Sangati

*Sangati*, the Hindi term for *consistency*, is a command line tool to makes it easier to check for logical inconsistencies or business logic errors in your data. It does so by automatically executing a series of SQL statements (provided in a JSON-formatted configuration file) that each output a single _count_ as a result. This result is then compared to either static values or the output of another SQL statement using logical operators that are also specified in the configuration file. 

While ideally, these checks would be implemented as constraints on the respective tables in an RDBMS or as descriptive relationships in the ORM of choice, frequently it is not practical to do so. For example, the code might have been ported from a legacy system that did not use an ORM or the data might be sourced from an ETL process that does not respect the constraints of the source database.

Currently, *Sangati* only supports connections to PostgreSQL and its variants. Support for other database systems will be added in the future.

## Usage

## Examples

## Frequently Asked Questions



