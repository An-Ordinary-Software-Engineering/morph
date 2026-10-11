package types

import "fmt"

type Primitive interface {
	Get() string
}

type Integer struct{}

func (i Integer) Get() string {
	return "INT"
}

type VarChar struct{
	Length uint8
}

func (s VarChar) Get() string {
	return fmt.Sprintf("VARCHAR(%d)", s.Length)
}

type Char struct{
	Length uint8
}

func (s Char) Get() string {
	return fmt.Sprintf("CHAR(%d)", s.Length)
}

type Timestamp struct{}

func (t Timestamp) Get() string {
	return "TIMESTAMP"
}