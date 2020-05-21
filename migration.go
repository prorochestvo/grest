// https://habr.com/ru/company/oleg-bunin/blog/461935/
package grest

import (
	"database/sql"
	"fmt"
	"grest/db"
	"grest/internal/dbase"
	"grest/internal/helper"
	"sort"
	"strings"
)

func NewMigration(driver db.Driver, router *Router) *migration {
	result := migration{}
	result.Table = "_migrations"
	result.router = router
	result.driver = driver
	return &result
}

type migration struct {
	Table  string
	driver db.Driver
	router *Router
}

func (this *migration) init() error {
	query := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ( version TEXT NOT NULL, apply_time INTEGER NOT NULL, PRIMARY KEY(version) );`, this.Table),
	}
	if err := this.driver.Exec(query...); err != nil {
		return err
	}
	return nil
}

func (this *migration) migrations() []db.Migration {
	result := make([]db.Migration, 0)
	for _, controller := range this.router.controllers {
		if controller == nil {
			continue
		}
		if c, ok := controller.(db.MigrationController); ok == true && c != nil {
			result = append(result, c.Migrations()...)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Version() < result[j].Version() })
	return result
}

func (this *migration) Version() string {
	if err := this.init(); err != nil {
		return ""
	}
	result := ""
	if ver, err := dbase.NewMigration(this.driver, this.Table).Version(); err != nil {
		_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	} else if len(ver) > 0 {
		result = ver
	}
	return result
}

/* exists migrations */
func (this *migration) Check() map[string]bool {
	if err := this.init(); err != nil {
		return make(map[string]bool, 0)
	}
	migrations := this.migrations()
	result := make(map[string]bool, len(migrations))
	for _, m := range migrations {
		result[m.Version()] = false
	}
	if history, err := dbase.NewMigration(this.driver, this.Table).History(); err != nil {
		_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	} else if history != nil {
		for _, version := range history {
			result[version] = false
		}
	}
	return result
}

/* check fields and migration, return error is not exists */
func (this *migration) Revise() error {
	if err := this.init(); err != nil {
		return err
	}
	result := make([]string, 0)
	// check migrations
	if history, err := dbase.NewMigration(this.driver, this.Table).History(); err != nil {
		return err
	} else if history != nil {
		migrations := this.migrations()
		for _, m := range migrations {
			version := m.Version()
			if helper.StringsIndexOf(history, version) < 0 {
				result = append(result, fmt.Sprintf("migration \"%s\" does not exist", version))
			}
		}
	}
	// check fields
	for _, controller := range this.router.controllers {
		if controller == nil {
			continue
		}
		model := getControllerModel(controller)
		if model == nil {
			continue
		}
		if fields := model.Fields(); fields != nil && len(fields) > 0 {
			wrong := make([]string, 0)
			for _, field := range fields {
				if _, err := this.driver.Select(db.NewSQLTable(model.Table()), []db.SQLField{db.NewSQLField(field.Name(), nil)}, nil, nil, nil, nil, db.NewSQLLimit(1), nil); err != nil && err != sql.ErrNoRows {
					wrong = append(wrong, field.Name())
				}
			}
			if wrong != nil && len(wrong) > 0 {
				column := "column"
				if len(wrong) > 1 {
					column = "columns"
				}
				sort.Strings(wrong)
				result = append(result, fmt.Sprintf("%s \"%s\" does not exist in the \"%s\" table", column, strings.Join(wrong, "\", \""), model.Table()))
			}
		}
	}
	// return all error
	if result != nil && len(result) > 0 {
		return fmt.Errorf(strings.Join(result, "\n"))
	}
	return nil
}

func (this *migration) Up() error {
	if err := this.init(); err != nil {
		return err
	}
	result := make([]string, 0)
	migrations := this.migrations()
	if history, err := dbase.NewMigration(this.driver, this.Table).History(); err != nil {
		return err
	} else {
		for _, m := range migrations {
			version := m.Version()
			if helper.StringsIndexOf(history, version) < 0 {
				if err := dbase.NewMigration(this.driver, this.Table).Append(version, m.Up()); err != nil {
					if this.router.Stderr != nil {
						_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("%s\n%s\n", m.Up(), err.Error())))
					}
					if this.router.Stdout != nil {
						_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("⚠ %s\n\t%s\n", version, strings.ReplaceAll(err.Error(), "\n", " "))))
					}
					result = append(result, err.Error())
				} else if this.router.Stdout != nil {
					_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("✔ %s\n", version)))
				}
			} else if this.router.Stdout != nil {
				_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("✔ %s\n", version)))
			}
		}
	}
	if len(result) > 0 {
		return fmt.Errorf(strings.Join(result, "\n"))
	}
	return nil
}

func (this *migration) Down() error {
	if err := this.init(); err != nil {
		return err
	}
	result := make([]string, 0)
	migrations := this.migrations()
	if current, err := dbase.NewMigration(this.driver, this.Table).Version(); err != nil {
		return err
	} else {
		for _, m := range migrations {
			version := m.Version()
			if current == version {
				if err = dbase.NewMigration(this.driver, this.Table).Remove(version, m.Down()); err != nil {
					if this.router.Stderr != nil {
						_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("%s\n%s\n", m.Down(), err.Error())))
					}
					if this.router.Stdout != nil {
						_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("⚠ %s\n\t%s\n", version, strings.ReplaceAll(err.Error(), "\n", " "))))
					}
					result = append(result, err.Error())
				} else if this.router.Stdout != nil {
					_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("✘ %s\n", m.Version())))
				}
				break
			}
		}
	}
	if len(result) > 0 {
		return fmt.Errorf(strings.Join(result, "\n"))
	}
	return nil
}

func (this *migration) Reset() error {
	if err := this.init(); err != nil {
		return err
	}
	result := make([]string, 0)
	migrations := this.migrations()
	if history, err := dbase.NewMigration(this.driver, this.Table).History(); err != nil {
		return err
	} else {
		for _, version := range history {
			var migration db.Migration = nil
			for _, m := range migrations {
				if version != m.Version() {
					continue
				}
				migration = m
				break
			}
			down := ""
			if migration != nil {
				down = migration.Down()
			}
			if err = this.driver.Exec(down, fmt.Sprintf(`DELETE FROM %s WHERE version = '%s';`, this.Table, version)); err != nil {
				if this.router.Stderr != nil {
					_, _ = this.router.Stderr.Write([]byte(fmt.Sprintf("%s\n%s\n", down, err.Error())))
				}
				if this.router.Stdout != nil {
					_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("⚠ %s\n\t%s\n", version, strings.ReplaceAll(err.Error(), "\n", " "))))
				}
				result = append(result, err.Error())
			} else if this.router.Stdout != nil {
				_, _ = this.router.Stdout.Write([]byte(fmt.Sprintf("✘ %s\n", version)))
			}
		}
	}
	if len(result) > 0 {
		return fmt.Errorf(strings.Join(result, "\n"))
	}
	return nil
}
