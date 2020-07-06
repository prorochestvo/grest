package db

type SQLTable interface {
	Name() string
}

func NewSQLTable(name string) SQLTable {
	result := sqlTable{}
	result.name = name
	return &result
}

type sqlTable struct {
	name string
}

func (this *sqlTable) Name() string {
	return this.name
}
