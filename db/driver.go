package db

type Driver interface {
	Select(table SQLTable, fields []SQLField, where []SQLWhere, groupBy []SQLGroupBy, having []SQLHaving, orderBy []SQLOrderBy, limit SQLLimit, offset SQLOffset) (rows []map[string]interface{}, err error)
	Insert(table SQLTable, fields []SQLField) (res interface{}, err error)
	Update(table SQLTable, fields []SQLField, where []SQLWhere) (err error)
	Delete(table SQLTable, where []SQLWhere) (err error)

	Exec(query ...string) error
}
