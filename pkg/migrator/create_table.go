package migrator

import (
	"fmt"

	"github.com/An-Ordinary-Software-Engineering/morph/pkg/constraints/table"
)

func CreateTable(name string, blueprint func(table *table.CreateTableConstraint) *table.CreateTableConstraint, options table.Options) *table.CreateTableConstraint {
	table := table.NewCreateTableConstraint(name)

	if options.DropIfExists {
		fmt.Fprintf(&table.Builder, "DROP TABLE IF EXISTS %s;\n\n", name)
	}

	fmt.Fprintf(&table.Builder, "CREATE TABLE %s (\n", name)
	table = blueprint(table)
	table.Build()

	if options.Engine != "" {
		fmt.Fprintf(&table.Builder, ") ENGINE=%s;\n", options.Engine)
	} else {
		fmt.Fprintf(&table.Builder, ");\n")
	}
	return table
}
