package db

type SQLOffset interface {
	Rows() int64
}

func NewSQLOffset(value int64) SQLOffset {
	result := sqlOffset{}
	result.value = value
	return &result
}

type sqlOffset struct {
	value int64
}

func (this *sqlOffset) Rows() int64 {
	return this.value
}
