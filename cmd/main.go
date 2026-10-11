package main

import (
	"fmt"

	"github.com/An-Ordinary-Software-Engineering/morph/pkg/constraints/table"
	"github.com/An-Ordinary-Software-Engineering/morph/pkg/types"
	"github.com/An-Ordinary-Software-Engineering/morph/pkg/migrator"
)

func main() {
	constraint := migrator.CreateTable(
		"user",
		func(table *table.CreateTableConstraint) *table.CreateTableConstraint {
			return table.
				Column("id", types.Integer{}).PrimaryKey().AutoIncrement().
				Column("name", types.VarChar{Length: 50}).NotNull().
				Column("nickname", types.VarChar{Length: 50}).NotNull().Unique().
				Column("email", types.VarChar{Length: 50}).NotNull().Unique().
				Column("password", types.VarChar{Length: 255}).NotNull().
				Column("created_on", types.Timestamp{}).Default(types.Defaults.CurrentTimestamp).NotNull()
		},
		table.Options{
			DropIfExists: true,
			Engine:       "INNODB",
		},
	)

	fmt.Println(constraint.Builder.String())
}
