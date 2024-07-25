package rok

import "github.com/lamlv2305/rok/postgres"

var (
	// Postgres is a global instance of the Postgres struct -> easy access on library
	Postgres = postgres.Postgres{}
)
