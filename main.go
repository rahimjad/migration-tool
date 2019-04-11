package main

import (
	"./migrator"
	"./postgres"
)

func main() {
	postgres.DropSchema()
	postgres.CreateSchema()
	migrator.MigrateUp()
}
