package dba

type Database struct {
	Name   string
	Tables []Table
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name string
	Type string
}
