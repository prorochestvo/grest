package db

type SQLLimit interface {
	Count() int64
}

func NewSQLLimit(value int64) SQLLimit {
	result := sqlLimit{}
	result.value = value
	return &result
}

type sqlLimit struct {
	value int64
}

func (this *sqlLimit) Count() int64 {
	return this.value
}
