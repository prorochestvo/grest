package db

type SQLGroupBy interface {
	Field() string
}

func NewSQLGroupBy(field string) SQLGroupBy {
	result := sqlGroupBy{}
	result.field = field
	return &result
}

type sqlGroupBy struct {
	field string
}

func (this *sqlGroupBy) Field() string {
	return this.field
}
