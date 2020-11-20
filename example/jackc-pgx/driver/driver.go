package driver

import (
  "database/sql"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/prorochestvo/grest/db"
	"strings"
)

// PostgreSQL driver
func Driver(db *pgx.ConnPool) db.Driver {
	result := driver{}
	result.ConnPool = db
	return &result
}

type driver struct {
	*pgx.ConnPool
}

func (this *driver) Select(table db.SQLTable, fields []db.SQLField, where []db.SQLWhere, groupBy []db.SQLGroupBy, having []db.SQLHaving, orderBy []db.SQLOrderBy, limit db.SQLLimit, offset db.SQLOffset) ([]map[string]interface{}, error) {
	query := db.NewSQLLinker().Select(table, fields, where, groupBy, having, orderBy, limit, offset)
	if rows, err := this.ConnPool.Query(query); err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if rows != nil {
		result := make([]map[string]interface{}, 0)
		defer rows.Close()
		if columns := rows.FieldDescriptions(); columns != nil && len(columns) > 0 {
			for rows.Next() {
				var values []interface{} = nil
				if values, err = rows.Values(); err != nil {
					return nil, err
				} else if values == nil {
					return nil, fmt.Errorf("empty dataset")
				}
				item := make(map[string]interface{}, len(columns))
				for i, col := range columns {
					item[col.Name] = values[i]
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
	if pos := strings.LastIndex(query, ";"); pos > 0 {
		query = fmt.Sprintf("%s\nRETURNING *;", query[:pos])
	}
	if rows, err := this.ConnPool.Query(query); err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if rows != nil {
		var result map[string]interface{} = nil
		defer rows.Close()
		if columns := rows.FieldDescriptions(); columns != nil && len(columns) > 0 {
			for rows.Next() {
				var values []interface{} = nil
				if values, err = rows.Values(); err != nil {
					return nil, err
				} else if values == nil {
					return nil, fmt.Errorf("empty dataset")
				}
				item := make(map[string]interface{}, len(columns))
				for i, col := range columns {
					item[col.Name] = values[i]
				}
				result = item
			}
		}
		return result, nil
	}
	return 0, nil
}

func (this *driver) Update(table db.SQLTable, fields []db.SQLField, where []db.SQLWhere) error {
	query := db.NewSQLLinker().Update(table, fields, where)
	if _, err := this.ConnPool.Exec(query); err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (this *driver) Delete(table db.SQLTable, where []db.SQLWhere) error {
	query := db.NewSQLLinker().Delete(table, where)
	if _, err := this.ConnPool.Exec(query); err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (this *driver) Exec(query ...string) error {
	if t, err := this.ConnPool.Begin(); err != nil {
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

func (this *driver) Escape(value string) string {
	return fmt.Sprintf(`"%s"`, value)
}
