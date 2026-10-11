package table

import (
	"fmt"
	"strings"

	"github.com/An-Ordinary-Software-Engineering/morph/pkg/constraints/columns"
	"github.com/An-Ordinary-Software-Engineering/morph/pkg/types"
)

type CreateTableConstraint struct {
	TableName string
    Builder strings.Builder
    payload  []columns.ColumnConstraint
    columnTarget int
}

func NewCreateTableConstraint(tableName string) *CreateTableConstraint {
    return &CreateTableConstraint{
        TableName: tableName,
        payload:  []columns.ColumnConstraint{},
        columnTarget: -1,
    }
}

func (t *CreateTableConstraint) Column(name string, typeField types.Primitive) *CreateTableConstraint {
    t.columnTarget++

    t.payload = append(t.payload, columns.ColumnConstraint{
        Name: name,
        Type: typeField.Get(),
        Constraints: []string{},
    })
	return t
}

func (t *CreateTableConstraint) PrimaryKey() *CreateTableConstraint {
    t.payload[t.columnTarget].Constraints = append(t.payload[t.columnTarget].Constraints, "PRIMARY KEY")
    return t
}

func (t *CreateTableConstraint) AutoIncrement() *CreateTableConstraint {
    t.payload[t.columnTarget].Constraints = append(t.payload[t.columnTarget].Constraints, "AUTO_INCREMENT")
    return t
}

func (t *CreateTableConstraint) NotNull() *CreateTableConstraint {
    t.payload[t.columnTarget].Constraints = append(t.payload[t.columnTarget].Constraints, "NOT NULL")
    return t
}

func (t *CreateTableConstraint) Unique() *CreateTableConstraint {
    t.payload[t.columnTarget].Constraints = append(t.payload[t.columnTarget].Constraints, "UNIQUE")
    return t
}

func (t *CreateTableConstraint) Default(value string) *CreateTableConstraint {
    t.payload[t.columnTarget].Constraints = append(t.payload[t.columnTarget].Constraints, "DEFAULT " + value)
    return t
}

func (t *CreateTableConstraint) Build() {
    for i, column := range t.payload {
        fmt.Fprintf(&t.Builder, "    %s %s", column.Name, column.Type)
        if len(column.Constraints) > 0 {
            fmt.Fprintf(&t.Builder, " %s", strings.Join(column.Constraints, " "))
        }
        if i < len(t.payload)-1 {
            fmt.Fprintf(&t.Builder, ",\n")
        } else {
            fmt.Fprintf(&t.Builder, "\n")
        }
    }
}
