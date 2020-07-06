package driver

import (
  "database/sql"
	"fmt"
	"github.com/prorochestvo/grest/db"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDriver(t *testing.T) {
	// sqlite database
	path := fmt.Sprintf("%s/86dac74fb580016230c1bf68a526b46d.db", os.TempDir())
	defer func() {
		_ = os.Remove(path)
	}()
	_ = os.Remove(path)
	base, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("db[sqlite]: %s", err.Error())
	} else if _, err := base.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"; CREATE TABLE IF NOT EXISTS "%s" ("id" INTEGER PRIMARY KEY AUTOINCREMENT, "text" TEXT);`, "users", "users")); err != nil {
		t.Fatalf("db[sqlite]: %s", err.Error())
	}
	driver := Driver(base)
	// exec
	if err := driver.Exec(fmt.Sprintf(`INSERT INTO "%s" ("text") VALUES('test-1'),('test-2'),('test-3');`, "users")); err != nil {
		t.Errorf("exec: %s", err.Error())
	}
	// select
	if rows, err := driver.Select(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, nil, nil, nil, []db.SQLOrderBy{db.NewSQLOrderBy("id", "ASC")}, nil, nil); err != nil {
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
	if res, err := driver.Insert(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("text", "test-4")}); err != nil {
		t.Errorf("insert: %s", err.Error())
	} else if id, ok := res.(int64); !ok || id != 4 {
		t.Errorf("insert: %s", "wrong field «id»")
	} else if rows, err := driver.Select(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("text", "test-4")}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("insert: %s", err.Error())
	} else if rows == nil || len(rows) != 1 {
		t.Errorf("insert: %s", "wrong rows count")
	} else if tmpId, ok := rows[0]["id"]; !ok {
		t.Errorf("insert: %s", "not found field «id»")
	} else if id, ok := tmpId.(int64); !ok || id != 4 {
		t.Errorf("insert: %s", "wrong field «id»")
	}
	// update
	if err := driver.Update(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("text", "test-----4")}, []db.SQLWhere{db.NewSQLWhere("id", 4)}); err != nil {
		t.Errorf("update: %s", err.Error())
	} else if rows, err := driver.Select(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("id", 4)}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("update: %s", err.Error())
	} else if rows == nil || len(rows) != 1 {
		t.Errorf("update: %s", "wrong rows count")
	} else if tmpText, ok := rows[0]["text"]; !ok {
		t.Errorf("update: %s", "not found field «text»")
	} else if text, ok := tmpText.(string); !ok || text != "test-----4" {
		t.Errorf("update: %s", "wrong field «text»")
	}
	// delete
	if err := driver.Delete(db.NewSQLTable("users"), []db.SQLWhere{db.NewSQLWhere("id", 4)}); err != nil {
		t.Errorf("delete: %s", err.Error())
	} else if rows, err := driver.Select(db.NewSQLTable("users"), []db.SQLField{db.NewSQLField("id", nil), db.NewSQLField("text", nil)}, []db.SQLWhere{db.NewSQLWhere("id", 4)}, nil, nil, nil, nil, nil); err != nil {
		t.Errorf("delete: %s", err.Error())
	} else if rows == nil || len(rows) != 0 {
		t.Errorf("delete: %s", "wrong rows count")
	}
}
