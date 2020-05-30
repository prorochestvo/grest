package jackc_pgx

import (
	"fmt"
	"github.com/jackc/pgx"
	"grest/db"
	"os"
	"strconv"
	"testing"
)

const (
	PSQLHost string = "127.0.0.1"
	PSQLPort uint16 = 5432
	PSQLBase string = "test"
	PSQLUser string = "tester"
	PSQLPass string = "password"
)

func TestDriver(t *testing.T) {
	// pgx config
	psqlConfig := pgx.ConnConfig{Host: PSQLHost, Port: PSQLPort, User: PSQLUser, Password: PSQLPass, Database: PSQLBase}
	if host := os.Getenv("DB_HOST"); len(host) > 0 {
		psqlConfig.Host = host
	}
	if port, err := strconv.ParseUint(os.Getenv("DB_PORT"), 10, 16); err == nil {
		psqlConfig.Port = uint16(port)
	}
	if base := os.Getenv("DB_BASE"); len(base) > 0 {
		psqlConfig.Database = base
	}
	if user := os.Getenv("DB_USER"); len(user) > 0 {
		psqlConfig.User = user
	}
	if pass := os.Getenv("DB_PASS"); len(pass) > 0 {
		psqlConfig.Password = pass
	}
	// postgres database
	base, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: psqlConfig})
	if err != nil {
		t.Fatalf("db[postgresql]: %s", err.Error())
	} else if _, err := base.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"; CREATE TABLE IF NOT EXISTS "%s" ("id" BIGSERIAL NOT NULL, "text" TEXT, PRIMARY KEY ("id"));`, "_driver_clients", "_driver_clients")); err != nil {
		base.Close()
		t.Fatalf("db[postgresql]: %s", err.Error())
	}
	defer base.Close()
	driver := Driver(base)
	// exec
	if err := driver.Exec(fmt.Sprintf(`INSERT INTO "%s" ("text") VALUES('test-1'),('test-2'),('test-3');`, "_driver_clients")); err != nil {
		t.Errorf("exec: %s", err.Error())
	}
	// select
	if rows, err := driver.Select(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, nil, nil, nil, []db.SQLOrderBy{db.NewSQLOrderBy("id", "ASC")}, nil, nil); err != nil {
		t.Errorf("select: %s", err.Error())
	} else if rows == nil || len(rows) != 3 {
		t.Errorf("select: %s", "wrong rows count")
	} else if tmpId, ok := rows[2]["id"]; !ok {
		t.Errorf("select: %s", "not found field «id»")
	} else if id, ok := tmpId.(int64); !ok || id != 3 {
		t.Errorf("select: %s", "wrong field «id»")
	} else if tmpText, ok := rows[2]["text"]; !ok {
		t.Errorf("select: %s", "not found field «text»")
	} else if text, ok := tmpText.(string); !ok || text != "test-3" {
		t.Errorf("select: %s", "wrong field «text»")
	}
	// insert
	if res, err := driver.Insert(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("text", "test-4")}); err != nil {
		t.Errorf("insert: %s", err.Error())
	} else if m, ok := res.(map[string]interface{}); !ok || m == nil {
		t.Errorf("insert: %s", "empty dataset")
	} else if tmp, ok := m["id"]; !ok || tmp == nil {
		t.Errorf("insert: %s", "wrong field «id»")
	} else if id, ok := tmp.(int64); !ok || id != 4 {
		t.Errorf("insert: %s", "wrong field «id»")
	} else if rows, err := driver.Select(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("text", "test-4")}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("insert: %s", err.Error())
	} else if rows == nil || len(rows) != 1 {
		t.Errorf("insert: %s", "wrong rows count")
	} else if tmpId, ok := rows[0]["id"]; !ok {
		t.Errorf("insert: %s", "not found field «id»")
	} else if id, ok := tmpId.(int64); !ok || id != 4 {
		t.Errorf("insert: %s", "wrong field «id»")
	}
	// update
	if err := driver.Update(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("text", "test-----4")}, []db.SQLWhere{db.NewSQLWhere("id", 4)}); err != nil {
		t.Errorf("update: %s", err.Error())
	} else if rows, err := driver.Select(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("id", 4)}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("update: %s", err.Error())
	} else if rows == nil || len(rows) != 1 {
		t.Errorf("update: %s", "wrong rows count")
	} else if tmpText, ok := rows[0]["text"]; !ok {
		t.Errorf("update: %s", "not found field «text»")
	} else if text, ok := tmpText.(string); !ok || text != "test-----4" {
		t.Errorf("update: %s", "wrong field «text»")
	}
	// delete
	if err := driver.Delete(db.NewSQLTable("_driver_clients"), []db.SQLWhere{db.NewSQLWhere("id", 4)}); err != nil {
		t.Errorf("delete: %s", err.Error())
	} else if rows, err := driver.Select(db.NewSQLTable("_driver_clients"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("id", 4)}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("delete: %s", err.Error())
	} else if rows == nil || len(rows) != 0 {
		t.Errorf("delete: %s", "wrong rows count")
	}
}
