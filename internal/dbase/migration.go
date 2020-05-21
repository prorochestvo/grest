package dbase

import (
	"fmt"
	"grest/db"
	"time"
)

type Migration interface {
	Version() (string, error)
	History() ([]string, error)
	Exists(version string) (bool, error)
	Append(version, sql string) error
	Remove(version, sql string) error
}

func NewMigration(driver db.Driver, table string) Migration {
	result := migration{}
	result.driver = driver
	result.table = table
	return &result
}

type migration struct {
	table  string
	driver db.Driver
}

// текущая версии по времени реализации в бд
func (this *migration) Version() (string, error) {
	result := ""
	if rows, err := this.driver.Select(db.NewSQLTable(this.table), []db.SQLField{db.NewSQLField("version", nil)}, nil, nil, nil, []db.SQLOrderBy{db.NewSQLOrderBy("apply_time", "DESC"), db.NewSQLOrderBy("version", "DESC")}, db.NewSQLLimit(1), nil); err != nil {
		return "", err
	} else if rows != nil && len(rows) == 1 && rows[0] != nil {
		if v, ok := rows[0]["version"]; ok && v != nil {
			value, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("wrong field")
			}
			result = value
		} else {
			return "", fmt.Errorf("missing field")
		}
	}
	return result, nil
}

// список версии по времени реализации в бд (первый (0) является текущей версией)
func (this *migration) History() ([]string, error) {
	result := make([]string, 0)
	if rows, err := this.driver.Select(db.NewSQLTable(this.table), []db.SQLField{db.NewSQLField("version", nil)}, nil, nil, nil, []db.SQLOrderBy{db.NewSQLOrderBy("apply_time", "DESC"), db.NewSQLOrderBy("version", "DESC")}, nil, nil); err != nil {
		return make([]string, 0), err
	} else if rows != nil && len(rows) > 1 {
		for _, row := range rows {
			if v, ok := row["version"]; ok && v != nil {
				value, ok := v.(string)
				if !ok {
					return make([]string, 0), fmt.Errorf("wrong field")
				}
				result = append(result, value)
			} else {
				return make([]string, 0), fmt.Errorf("missing field")
			}
		}
	}
	return result, nil
}

// проверка наличия версии
func (this *migration) Exists(version string) (bool, error) {
	result := false
	if rows, err := this.driver.Select(db.NewSQLTable(this.table), []db.SQLField{db.NewSQLField("COUNT(apply_time) as cnt", nil)}, []db.SQLWhere{db.NewSQLWhere("version", version)}, nil, nil, nil, db.NewSQLLimit(1), nil); err != nil {
		return false, err
	} else if rows != nil && len(rows) > 1 {
		for _, row := range rows {
			if v, ok := row["cnt"]; ok && v != nil {
				value, ok := v.(int)
				result = ok && value == 1
				break
			}
		}
	}
	return result, nil
}

// добавить версию в хеш таблице
func (this *migration) Append(version, sql string) error {
	return this.driver.Exec(sql, fmt.Sprintf("INSERT INTO %s (version, apply_time) VALUES ('%s', %d);", this.table, version, time.Now().UnixNano()/int64(time.Second)))
}

// удалить версию в хеш таблице
func (this *migration) Remove(version, sql string) error {
	return this.driver.Exec(sql, fmt.Sprintf("DELETE FROM %s WHERE version = '%s';", this.table, version))
}
