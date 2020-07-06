package driver

import (
  "database/sql"
	"fmt"
	"github.com/prorochestvo/grest/db"
)

// SQLite driver
func Driver(db *sql.DB) db.Driver {
	result := driver{}
	result.DB = db
	return &result
}

type driver struct {
	*sql.DB
}

func (this *driver) Select(table db.SQLTable, fields []db.SQLField, where []db.SQLWhere, groupBy []db.SQLGroupBy, having []db.SQLHaving, orderBy []db.SQLOrderBy, limit db.SQLLimit, offset db.SQLOffset) ([]map[string]interface{}, error) {
	query := db.NewSQLLinker().Select(table, fields, where, groupBy, having, orderBy, limit, offset)
	if rows, err := this.DB.Query(query); err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if rows != nil {
		result := make([]map[string]interface{}, 0)
		defer rows.Close()
		if columns, err := rows.Columns(); err != nil {
			return nil, err
		} else if columns != nil && len(columns) > 0 {
			values := make([]interface{}, len(columns))
			for i, _ := range values {
				var tmp interface{}
				values[i] = &tmp
			}
			for rows.Next() {
				if err := rows.Scan(values...); err != nil {
					return nil, err
				}
				item := make(map[string]interface{}, len(columns))
				for i, col := range columns {
					item[col] = *values[i].(*interface{})
				}
				result = append(result, item)
			}
		}
		return result, nil
	}
	return make([]map[string]interface{}, 0), nil
}

func (this *driver) Insert(table db.SQLTable, fields []db.SQLField) (interface{}, error) {
	query := db.NewSQLLinker().Insert(table, fields)
	if res, err := this.DB.Exec(query); err != nil && err != sql.ErrNoRows {
		return 0, err
	} else if res == nil {
		return 0, fmt.Errorf("empty dataset")
	} else if id, err := res.LastInsertId(); err != nil {
		return 0, err
	} else if id > 0 {
		return id, nil
	}
	return 0, nil
}

func (this *driver) Update(table db.SQLTable, fields []db.SQLField, where []db.SQLWhere) error {
	query := db.NewSQLLinker().Update(table, fields, where)
	if _, err := this.DB.Exec(query); err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (this *driver) Delete(table db.SQLTable, where []db.SQLWhere) error {
	query := db.NewSQLLinker().Delete(table, where)
	if _, err := this.DB.Exec(query); err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (this *driver) Exec(query ...string) error {
	if t, err := this.DB.Begin(); err != nil {
		return err
	} else if t != nil {
		defer func() {
			if err != nil {
				_ = t.Rollback()
				return
			}
			err = t.Commit()
		}()
		for _, q := range query {
			if len(q) == 0 {
				continue
			} else if _, err = t.Exec(q); err != nil {
				return err
			}
		}
	}
	return nil
}
