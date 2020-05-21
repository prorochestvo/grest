package db

type SQLOrderBy interface {
	Field() string
	Sort() string
}

func NewSQLOrderBy(field string, sort string) SQLOrderBy {
	result := sqlOrderBy{}
	result.field = field
	result.sort = sort
	return &result
}

type sqlOrderBy struct {
	field string
	sort  string
}

func (this *sqlOrderBy) Field() string {
	return this.field
}

func (this *sqlOrderBy) Sort() string {
	return this.sort
}
