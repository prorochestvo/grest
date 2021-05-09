package main

import (
  "bytes"
  "context"
	"database/sql"
	"encoding/json"
	"fmt"
  "io/ioutil"
  "log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

  . "database_sql-example/driver"

  _ "github.com/mattn/go-sqlite3"
  "github.com/prorochestvo/grest"
  "github.com/prorochestvo/grest/db"
  "github.com/prorochestvo/grest/usr"
)

var (
	User            grest.ControllerWithModel = &ControllerUser{}
	UserSession     grest.ControllerWithModel = &ControllerUserSession{}
	OperatingSystem grest.ControllerWithModel = &ControllerOperatingSystem{}
	APIDocs         grest.Controller          = &ControllerAPIDocs{}
)

func TestJSONRouter(t *testing.T) {
	wg := sync.WaitGroup{}
	var tmp interface{} = nil
  var itmp int64 = 0
  var stmp string = ""
  // database connection
	path := fmt.Sprintf("%s/7008360c46c3baa97dacc27a7f994f77.db", os.TempDir())
	defer func() {
		_ = os.Remove(path)
	}()
	_ = os.Remove(path)
  dbase, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("grest[db]: %s", err.Error())
	}
	// router
	router := grest.NewJSONRouter(Driver(dbase))
	router.Migration.Table = "__migrations"
	router.Stderr = nil
	router.Stdout = nil
	router.Version = "1.0"
	if err := router.Listen(APIDocs); err != nil {
		t.Fatalf("grest: %s", err.Error())
	}
	if err := router.Listen(User, UserSession, OperatingSystem); err != nil {
		t.Fatalf("grest: %s", err.Error())
	}
	router.AccessControl.User = func(r *grest.Request) (usr.User, error) {
		role, _ := strconv.ParseInt(r.Header.Get("Authorization"), 10, 32)
		return usr.NewUser(r.Header.Get("Authorization"), usr.Role(role)), nil
	}
  // reset dbase
  if _, err := dbase.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";",router.Migration.Table)); err != nil {
    log.Fatalf("PSQL %s", err.Error())
  }
  if _, err := dbase.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";",User.Model().Table())); err != nil {
    log.Fatalf("PSQL %s", err.Error())
  }
  if _, err := dbase.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";",UserSession.Model().Table())); err != nil {
    log.Fatalf("PSQL %s", err.Error())
  }
  if _, err := dbase.Exec(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";",OperatingSystem.Model().Table())); err != nil {
    log.Fatalf("PSQL %s", err.Error())
  }
	// server
	server := http.Server{}
	server.Handler = router
	server.Addr = fmt.Sprintf(":%d", HTTPPort)
	server.WriteTimeout = 30 * time.Second
	server.ReadTimeout = 30 * time.Second
	go func() {
		wg.Add(1)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("grest[server]: %s", err.Error())
		}
		wg.Done()
	}()
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf("grest[server]: %s", err.Error())
		}
		wg.Wait()
	}()
	time.Sleep(500 * time.Millisecond)
	// check migration (up | down | reset)
	if version := router.Migration.Version(); len(version) > 0 {
		t.Errorf("grest[migration]: wrong start version (%s)", version)
	} else if err := router.Migration.Up(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", User.Model().Table())).Scan(&itmp); err != nil && err != sql.ErrNoRows || itmp == 0 {
		t.Errorf("grest[migration]: %s (%s). %s", "not exists table", User.Model().Table(), err.Error())
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", UserSession.Model().Table())).Scan(&itmp); err != nil && err != sql.ErrNoRows || itmp == 0 {
		t.Errorf("grest[migration]: %s (%s). %s", "not exists table", UserSession.Model().Table(), err.Error())
	} else if err := router.Migration.Down(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if err := router.Migration.Down(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", UserSession.Model().Table())).Scan(&itmp); err == nil {
		t.Errorf("grest[migration]: %s (%s)", "must no such table", UserSession.Model().Table())
	} else if err := router.Migration.Up(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if err := router.Migration.Reset(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if err := router.Migration.Up(); err != nil {
		t.Errorf("grest[migration]: %s", err.Error())
	} else if version := router.Migration.Version(); version != "v-user_session-0002" {
		t.Errorf("grest[migration]: wrong last version (%s)", version)
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", User.Model().Table())).Scan(&itmp); err != nil && err != sql.ErrNoRows || itmp == 0 {
		t.Errorf("grest[migration]: %s (%s). %s", "not exists table", User.Model().Table(), err.Error())
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", UserSession.Model().Table())).Scan(&itmp); err != nil && err != sql.ErrNoRows || itmp == 0 {
		t.Errorf("grest[migration]: %s (%s). %s", "not exists table", UserSession.Model().Table(), err.Error())
	} else {
    defer func() {
      if err := router.Migration.Reset(); err != nil {
        log.Fatalf("grest[migration]: %s", err.Error())
      }
    }()
  }
	// check controller role assess
	if code, _, _, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/controller/list", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[role-assess]: %s", err.Error())
	} else if code != http.StatusForbidden {
		t.Errorf("grest[role-assess]: wrong response status (%d)", code)
	} else if code, _, _, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/controller/list", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[role-assess]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[role-assess]: wrong response status (%d)", code)
	}
	// check model fields role assess
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user/3", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[role-assess]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[role-assess]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[role-assess]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[role-assess]: wrong response. %s", err.Error())
	} else if item, ok := tmp.(map[string]interface{}); !ok || item == nil {
		t.Errorf("grest[role-assess]: wrong response. %s.", "json object must be 'map'")
	} else if _, ok := item["password"]; ok {
		t.Errorf("grest[role-assess]: private field exists ('password'). (%q).", item)
	} else if code, _, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://127.0.0.1:%d/user/3", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, map[string]interface{}{
		"name":     "usr-34849e699ac449ab30b97c58974",
		"password": "pwd-34849e699ac449ab30b97c58974",
	}); err != nil {
		t.Errorf("grest[role-assess]: %s", err.Error())
	} else if code != http.StatusForbidden {
		t.Errorf("grest[role-assess]: wrong response status (%d) %s.", code, body)
	} else if code, _, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://127.0.0.1:%d/user/3", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"name":           "usr-3e699ac449ab30b97c484958974",
		"password":       "pwd-3e699ac449ab30b97c484958974",
		"internal-field": "fld-3e699ac449ab30b97c484958974",
	}); err != nil {
		t.Errorf("grest[role-assess]: %s", err.Error())
	} else if code != http.StatusUnprocessableEntity {
		t.Errorf("grest[role-assess]: wrong response status (%d) %s.", code, body)
	}
	// check model fields validator
	if code, _, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://127.0.0.1:%d/user/3", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"login": "usr-3e699ac449ab30b97c484958974",
	}); err != nil {
		t.Errorf("grest[fields-validator]: %s", err.Error())
	} else if code != http.StatusUnprocessableEntity {
		t.Errorf("grest[fields-validator]: wrong response status (%d) %s.", code, body)
	}
	// check pagination action
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user-session?:sort[ip]=ASC&page=2&page[size]=7", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[pagination-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[pagination-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[pagination-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[pagination-action]: wrong response. %s", err.Error())
	} else if res, ok := tmp.(map[string]interface{}); !ok || res == nil {
		t.Errorf("grest[pagination-action]: wrong response, %s (%s).", "json object must be 'map'", string(body))
	} else if data, ok := res["data"]; !ok || data == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found rows data")
	} else if items, ok := data.([]interface{}); !ok || items == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 7 {
		t.Errorf("grest[pagination-action]: wrong object list (%d).", len(items))
	} else if item, ok := items[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[pagination-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if iIP, ok := item["ip"]; !ok {
		t.Errorf("grest[pagination-action]: not found \"ip\" on object[0]. %s (%s).", "json object must be 'string'", item)
	} else if ip, ok := iIP.(string); !ok || ip != "1.0.0.008" {
		t.Errorf("grest[pagination-action]: wrong object[0][\"ip\"]. %s (%s).", "json object must be 'string'", ip)
	} else if meta, ok := res["meta"]; !ok || meta == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found meta data")
	} else if info, ok := meta.(map[string]interface{}); !ok || info == nil || len(info) == 0 {
		t.Errorf("grest[pagination-action]: wrong meta data. %s.", "json object must be 'map'")
	} else if iTotalEntries, ok := info["total_entries"]; !ok || iTotalEntries == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_entries'")
	} else if totalEntries, ok := iTotalEntries.(float64); !ok || totalEntries != 36.0 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_entries\"]. %s (%d).", "json object must be 'integer'", int64(totalEntries))
	} else if iCurrentPage, ok := info["current_page"]; !ok || iCurrentPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'current_page'")
	} else if currentPage, ok := iCurrentPage.(float64); !ok || currentPage != 2 {
		t.Errorf("grest[pagination-action]: wrong meta[\"current_page\"]. %s (%d).", "json object must be 'integer'", int64(currentPage))
	} else if iTotalPages, ok := info["total_pages"]; !ok || iTotalPages == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_pages'")
	} else if totalPages, ok := iTotalPages.(float64); !ok || totalPages != 6 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_pages\"]. %s (%d).", "json object must be 'integer'", int64(totalPages))
	} else if iPerPage, ok := info["per_page"]; !ok || iPerPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'per_page'")
	} else if perPage, ok := iPerPage.(float64); !ok || perPage != 7 {
		t.Errorf("grest[pagination-action]: wrong meta[\"per_page\"]. %s (%d).", "json object must be 'integer'", int64(perPage))
	} else if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user-session?1:sort[user_id]=DESC&2:sort[ip]=ASC&user_id=1", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[pagination-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[pagination-action]: wrong response status (%d) %s.", code, body)
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[pagination-action]: wrong response. %s", err.Error())
	} else if res, ok := tmp.(map[string]interface{}); !ok || res == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "json object must be 'map'")
	} else if data, ok := res["data"]; !ok || data == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found rows data")
	} else if items, ok := data.([]interface{}); !ok || items == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 10 {
		t.Errorf("grest[pagination-action]: wrong object list (%d).", len(items))
	} else if item, ok := items[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[pagination-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if iIP, ok := item["ip"]; !ok {
		t.Errorf("grest[pagination-action]: not found \"ip\" on object[0]. %s (%s).", "json object must be 'string'", item)
	} else if ip, ok := iIP.(string); !ok || ip != "1.0.0.001" {
		t.Errorf("grest[pagination-action]: wrong object[0][\"ip\"]. %s (%s).", "json object must be 'string'", ip)
	} else if item, ok := items[len(items)-1].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[pagination-action]: wrong object[%d]. %s.", len(items)-1, "json object must be 'map'")
	} else if iIP, ok := item["ip"]; !ok {
		t.Errorf("grest[pagination-action]: not found \"ip\" on object[%d]. %s (%s).", len(items)-1, "json object must be 'string'", item)
	} else if ip, ok := iIP.(string); !ok || ip != "1.0.0.010" {
		t.Errorf("grest[pagination-action]: wrong object[0][\"ip\"]. %s (%s).", "json object must be 'string'", ip)
	} else if meta, ok := res["meta"]; !ok || meta == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found meta data")
	} else if info, ok := meta.(map[string]interface{}); !ok || info == nil || len(info) == 0 {
		t.Errorf("grest[pagination-action]: wrong meta data. %s.", "json object must be 'map'")
	} else if iTotalEntries, ok := info["total_entries"]; !ok || iTotalEntries == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_entries'")
	} else if totalEntries, ok := iTotalEntries.(float64); !ok || totalEntries != 24.0 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_entries\"]. %s (%d).", "json object must be 'integer'", int64(totalEntries))
	} else if iCurrentPage, ok := info["current_page"]; !ok || iCurrentPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'current_page'")
	} else if currentPage, ok := iCurrentPage.(float64); !ok || currentPage != 1 {
		t.Errorf("grest[pagination-action]: wrong meta[\"current_page\"]. %s (%d).", "json object must be 'integer'", int64(currentPage))
	} else if iTotalPages, ok := info["total_pages"]; !ok || iTotalPages == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_pages'")
	} else if totalPages, ok := iTotalPages.(float64); !ok || totalPages != 3 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_pages\"]. %s (%d).", "json object must be 'integer'", int64(totalPages))
	} else if iPerPage, ok := info["per_page"]; !ok || iPerPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'per_page'")
	} else if perPage, ok := iPerPage.(float64); !ok || perPage != 10 {
		t.Errorf("grest[pagination-action]: wrong meta[\"per_page\"]. %s (%d).", "json object must be 'integer'", int64(perPage))
	} else if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user-session?email=login@mail.ru", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[pagination-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[pagination-action]: wrong response status (%d) %s.", code, body)
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[pagination-action]: wrong response. %s", err.Error())
	} else if res, ok := tmp.(map[string]interface{}); !ok || res == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "json object must be 'map'")
	} else if data, ok := res["data"]; !ok || data == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found rows data")
	} else if items, ok := data.([]interface{}); !ok || items == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 10 {
		t.Errorf("grest[pagination-action]: wrong object list (%d).", len(items))
	} else if meta, ok := res["meta"]; !ok || meta == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found meta data")
	} else if info, ok := meta.(map[string]interface{}); !ok || info == nil || len(info) == 0 {
		t.Errorf("grest[pagination-action]: wrong meta data. %s.", "json object must be 'map'")
	} else if iTotalEntries, ok := info["total_entries"]; !ok || iTotalEntries == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_entries'")
	} else if totalEntries, ok := iTotalEntries.(float64); !ok || totalEntries != 36.0 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_entries\"]. %s (%d).", "json object must be 'integer'", int64(totalEntries))
	} else if iCurrentPage, ok := info["current_page"]; !ok || iCurrentPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'current_page'")
	} else if currentPage, ok := iCurrentPage.(float64); !ok || currentPage != 1 {
		t.Errorf("grest[pagination-action]: wrong meta[\"current_page\"]. %s (%d).", "json object must be 'integer'", int64(currentPage))
	} else if iTotalPages, ok := info["total_pages"]; !ok || iTotalPages == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'total_pages'")
	} else if totalPages, ok := iTotalPages.(float64); !ok || totalPages != 4 {
		t.Errorf("grest[pagination-action]: wrong meta[\"total_pages\"]. %s (%d).", "json object must be 'integer'", int64(totalPages))
	} else if iPerPage, ok := info["per_page"]; !ok || iPerPage == nil {
		t.Errorf("grest[pagination-action]: wrong response. %s.", "not found 'per_page'")
	} else if perPage, ok := iPerPage.(float64); !ok || perPage != 10 {
		t.Errorf("grest[pagination-action]: wrong meta[\"per_page\"]. %s (%d).", "json object must be 'integer'", int64(perPage))
	}
	// check list action
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user?:sort[id]=ASC", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[list-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[list-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[list-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[list-action]: wrong response. %s", err.Error())
	} else if items, ok := tmp.([]interface{}); !ok || items == nil {
		t.Errorf("grest[list-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 3 {
		t.Errorf("grest[list-action]: wrong object list (%d).", len(items))
	} else if item, ok := items[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[list-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[list-action]: not found \"id\" on object[0]. %s (%s).", "json object must be 'integer'", item)
	} else if id, ok := iId.(float64); !ok || id != 1.0 {
		t.Errorf("grest[list-action]: wrong object[0][\"id\"]. %s (%d).", "json object must be 'integer'", int(id))
	} else if iSession, ok := item["session"]; !ok {
		t.Errorf("grest[list-action]: not found \"session\" on object[0]. %s (%s).", "json object must be 'list'", item)
	} else if session, ok := iSession.([]interface{}); !ok || session == nil {
		t.Errorf("grest[list-action]: wrong object[0][\"session\"]. %s (%s).", "json object must be 'list'", session)
	} else if len(session) != 24 {
		t.Errorf("grest[list-action]: wrong len object[0][\"session\"] (%d).", len(session))
	} else if item, ok := session[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[list-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if iOperatingSystem, ok := item["operating-system"]; !ok {
		t.Errorf("grest[list-action]: not found \"operating-system\" on object[0][\"session\"][\"operating-system\"]]. %s (%s).", "json object must be 'map'", item)
	} else if operatingSystem, ok := iOperatingSystem.(map[string]interface{}); !ok || operatingSystem == nil || len(operatingSystem) == 0 {
		t.Errorf("grest[list-action]: wrong object[0][\"session\"][\"operating-system\"]. %s (%s).", "json object must be 'map'", iOperatingSystem)
	} else if iDescription, ok := operatingSystem["description"]; !ok {
		t.Errorf("grest[list-action]: not found \"operating-system\" on object[0][\"session\"][\"description\"]. %s (%s).", "json object must be 'string'", operatingSystem)
	} else if description, ok := iDescription.(string); !ok || description != "LINUX-x86" {
		t.Errorf("grest[list-action]: wrong object[0][\"session\"][\"operating-system\"][\"description\"]. %s (%s).", "json object must be 'string'", description)
	} else if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user?login=support", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[list-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[list-action]: wrong response status (%d) %s.", code, body)
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[list-action]: wrong response. %s", err.Error())
	} else if items, ok := tmp.([]interface{}); !ok || items == nil {
		t.Errorf("grest[list-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 1 {
		t.Errorf("grest[list-action]: wrong object list (%d).", len(items))
	} else if item, ok := items[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[list-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if sLogin, ok := item["login"]; !ok {
		t.Errorf("grest[list-action]: not found \"id\" on object[0]. %s (%s).", "json object must be 'string'", item)
	} else if login, ok := sLogin.(string); !ok || login != "support" {
		t.Errorf("grest[list-action]: wrong object[0][\"login\"]. %s (%s).", "json object must be 'string'", login)
	} else if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user?email=login@mail.ru", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[list-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[list-action]: wrong response status (%d) %s.", code, body)
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[list-action]: wrong response. %s", err.Error())
	} else if items, ok := tmp.([]interface{}); !ok || items == nil {
		t.Errorf("grest[list-action]: wrong response. %s.", "json object must be 'list'")
	} else if len(items) != 3 {
		t.Errorf("grest[list-action]: wrong object list (%d).", len(items))
	} else if item, ok := items[0].(map[string]interface{}); !ok || item == nil || len(item) == 0 {
		t.Errorf("grest[list-action]: wrong object[0]. %s.", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[list-action]: not found \"id\" on object[0]. %s (%s).", "json object must be 'integer'", item)
	} else if id, ok := iId.(float64); !ok || id != 1.0 {
		t.Errorf("grest[list-action]: wrong object[0][\"id\"]. %s (%d).", "json object must be 'integer'", int(id))
	}
	// check view action
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user/2", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[view-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[view-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[view-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[view-action]: wrong response. %s", err.Error())
	} else if item, ok := tmp.(map[string]interface{}); !ok || item == nil {
		t.Errorf("grest[view-action]: wrong response. %s.", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[view-action]: not found \"id\" on object. %s (%s).", "json object must be 'integer'", item)
	} else if id, ok := iId.(float64); !ok || id != 2.0 {
		t.Errorf("grest[view-action]: wrong object[\"id\"]. %s (%d).", "json object must be 'integer'", int(id))
	}
	// check create action
	if code, head, body, err := httpQuery(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/user", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"login":    "user-login",
		"password": "user-ee4dd4983e0132201c7a73e207547429",
		"name":     "user-ee4dd4983e0132201c7a73e207547429",
	}); err != nil {
		t.Errorf("grest[create-action]: %s", err.Error())
	} else if code != http.StatusCreated {
		t.Errorf("grest[create-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[create-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[create-action]: wrong response, %s", err.Error())
	} else if item, ok := tmp.(map[string]interface{}); !ok || item == nil {
		t.Errorf("grest[create-action]: wrong response, %s", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[create-action]: not found \"id\" on object, %s (%s)", "json object must be 'integer'", body)
	} else if id, ok := iId.(float64); !ok || id != 4.0 {
		t.Errorf("grest[create-action]: wrong object[\"id\"], %s (%d)", "json object must be 'integer'", int(id))
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE id = %d;", User.Model().Table(), 4)).Scan(&itmp); err != nil {
		t.Errorf("grest[create-action]: %s", err.Error())
	} else if itmp != 1 {
		t.Errorf("grest[create-action]: not found new user(#4).")
	} else if code, _, body, err := httpQuery(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/user", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"login":    "user-login",
		"password": "user-ee4dd4983e0132201c7a73e207547429",
		"name":     "user-ee4dd4983e0132201c7a73e207547429",
		"email":    "email@mail.ru",
	}); err != nil {
		t.Errorf("grest[create-action]: %s", err.Error())
	} else if code == http.StatusCreated {
		t.Errorf("grest[create-action]: wrong response status (%d) %s.", code, body)
	}
	// check update action
	if code, head, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://127.0.0.1:%d/user/4", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"name": "user-e2075474294983e013ee4dd2201c7a73",
	}); err != nil {
		t.Errorf("grest[update-action]: %s", err.Error())
	} else if code != http.StatusAccepted {
		t.Errorf("grest[update-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[update-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[update-action]: wrong response, %s", err.Error())
	} else if item, ok := tmp.(map[string]interface{}); !ok || item == nil {
		t.Errorf("grest[update-action]: wrong response, %s", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[update-action]: not found \"id\" on object, %s (%s)", "json object must be 'integer'", body)
	} else if id, ok := iId.(float64); !ok || id != 4.0 {
		t.Errorf("grest[update-action]: wrong object[\"id\"], %s (%d)", "json object must be 'integer'", int(id))
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT name FROM %s WHERE id = %d LIMIT 1;", User.Model().Table(), 4)).Scan(&stmp); err != nil {
		t.Errorf("grest[update-action]: %s", err.Error())
	} else if stmp != "user-e2075474294983e013ee4dd2201c7a73" {
		t.Errorf("grest[update-action]: wrong user(#4) name `%s`", stmp)
	} else if code, _, body, err := httpQuery(http.MethodPut, fmt.Sprintf("http://127.0.0.1:%d/user/4", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, map[string]interface{}{
		"name":  "user-e2075474294983e013ee4dd2201c7a73",
		"email": "email@mail.ru",
	}); err != nil {
		t.Errorf("grest[update-action]: %s", err.Error())
	} else if code == http.StatusAccepted {
		t.Errorf("grest[update-action]: wrong response status (%d) %s.", code, body)
	}
	// check delete action
	if code, head, body, err := httpQuery(http.MethodDelete, fmt.Sprintf("http://127.0.0.1:%d/user/4", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[delete-action]: %s", err.Error())
	} else if code != http.StatusAccepted {
		t.Errorf("grest[delete-action]: wrong response status (%d) %s.", code, body)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[delete-action]: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[delete-action]: wrong response, %s", err.Error())
	} else if item, ok := tmp.(map[string]interface{}); !ok || item == nil {
		t.Errorf("grest[delete-action]: wrong response, %s", "json object must be 'map'")
	} else if iId, ok := item["id"]; !ok {
		t.Errorf("grest[delete-action]: not found \"id\" on object, %s (%s)", "json object must be 'integer'", body)
	} else if id, ok := iId.(float64); !ok || id != 4.0 {
		t.Errorf("grest[delete-action]: wrong object[\"id\"], %s (%d)", "json object must be 'integer'", int(id))
	} else if err := dbase.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = %d;", User.Model().Table(), 4)).Scan(&itmp); err != nil {
		t.Errorf("grest[delete-action]: %s", err.Error())
	} else if itmp != 0 {
		t.Errorf("grest[delete-action]: user(#4) exists")
	}
	// check custom action
	if code, _, body, err := httpQuery(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/user/token", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[custom-action]: %s", err.Error())
	} else if code != http.StatusForbidden {
		t.Errorf("grest[custom-action]: wrong response status (%d) %s.", code, body)
	} else if code, _, body, err := httpQuery(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/user/token", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleUser)}, nil); err != nil {
		t.Errorf("grest[custom-action]: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[custom-action]: wrong response status (%d) %s.", code, body)
	} else if s := string(body); s != "\"user.Token\"" {
		t.Errorf("grest[view-action]: wrong response (%s).", s)
	}
	// check module controllers-share
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/controller/list", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[module-share-controllers]: modules: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[module-share-controllers]: modules: wrong response status (%d)", code)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "text/html") < 0 {
		t.Errorf("grest[module-share-controllers]: modules: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if strings.Index(string(body), "html>") < 0 || strings.Index(string(body), "body>") < 0 || strings.Index(string(body), "head>") < 0 {
		t.Errorf("grest[module-share-controllers]: modules: %s", "not found html tags")
	}
	// check modules sql-editor
	tmp = "?uid=1&7:!in[id][]]=7&7:!in[id][]]=6&7:!in[id][]]=5&1:!between[ID][]=2&1:!between[ID][]=3&3:sort[id]=DESC&4:like[name]=admin&5:!like[name]=support&6:is_null[is_enabled]&:offset=123&:limit=1"
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/sql/editor/%s%s", HTTPPort, UserSession.Model().Table(), tmp), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[module-sql-editor]: modules: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[module-sql-editor]: modules: wrong response status (%d)", code)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "text/plain") < 0 {
		t.Errorf("grest[module-sql-editor]: modules: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if s := fmt.Sprintf("SELECT *\nFROM %s\n%s;", UserSession.Model().Table(), "WHERE (NOT(ID BETWEEN '2' AND '3')) AND (name LIKE 'admin') AND (name LIKE 'support') AND (IS NULL is_enabled) AND (NOT(id IN ('7', '6', '5'))) AND (uid = '1')\nORDER BY id DESC\nLIMIT 1\nOFFSET 123"); string(body) != s {
		t.Errorf("grest[module-sql-editor]: modules: wrong sql response (%s)\n\n%s", string(body), string(s))
	} else
  if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/sql/editor/%s%s", HTTPPort, User.Model().Table(), "?:!in[id][]=1&:!in[id][]=2&:!in[id][]=3"), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
    t.Errorf("grest[module-sql-editor]: modules: %s", err.Error())
  } else if code != http.StatusOK || string(body) != fmt.Sprintf("SELECT *\nFROM %s\nWHERE (NOT(id IN ('1', '2', '3')));", User.Model().Table()) {
    t.Errorf("grest[module-sql-editor]: modules: wrong response code (%d) or response body (%s)", code, body)
  } else
  if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/sql/editor/%s%s", HTTPPort, User.Model().Table(), "?2:|like[name]=A%25&1:like[name]=B%25"), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
    t.Errorf("grest[module-sql-editor]: modules: %s", err.Error())
  } else if code != http.StatusOK || string(body) != fmt.Sprintf("SELECT *\nFROM %s\nWHERE (name LIKE 'B%s') OR (name LIKE 'A%s');", User.Model().Table(), "%", "%") {
    t.Errorf("grest[module-sql-editor]: modules: wrong response code (%d) or response body (%s)", code, body)
  } else
  if code, _, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/sql/editor/%s%s", HTTPPort, User.Model().Table(), "?:sort[name]=ASC&role=admin"), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
    t.Errorf("grest[module-sql-editor]: modules: %s", err.Error())
  } else if code != http.StatusOK || string(body) != fmt.Sprintf("SELECT *\nFROM %s\nWHERE (role = 'admin')\nORDER BY name ASC;", User.Model().Table()) {
    t.Errorf("grest[module-sql-editor]: modules: wrong response code (%d) or response body (%s)", code, body)
  }
	// check modules http-parser
	if code, head, body, err := httpQuery(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/docs/request/parser", HTTPPort), map[string]string{"Authorization": fmt.Sprintf("%d", RoleAdmin)}, nil); err != nil {
		t.Errorf("grest[module-http-parser]: modules: %s", err.Error())
	} else if code != http.StatusOK {
		t.Errorf("grest[module-http-parser]: modules: wrong response status (%d)", code)
	} else if h, ok := head["Content-Type"]; !ok || strings.Index(strings.Join(h, " "), "application/json") < 0 {
		t.Errorf("grest[module-http-parser]: modules: %s (%s)", "wrong content-type header", strings.Join(h, " "))
	} else if err := json.Unmarshal(body, &tmp); err != nil {
		t.Errorf("grest[module-http-parser]: modules: %s", err.Error())
	} else if data, ok := tmp.(map[string]interface{}); !ok {
		t.Errorf("grest[module-http-parser]: modules: wrong response (%s)", string(body))
	} else if _, ok := data["url"]; !ok {
		t.Errorf("grest[module-http-parser]: modules: not found url on response (%s)", string(body))
	} else if _, ok := data["headers"]; !ok {
		t.Errorf("grest[module-http-parser]: modules: not found headers on response (%s)", string(body))
	} else if _, ok := data["body"]; !ok {
		t.Errorf("grest[module-http-parser]: modules: not found body on response (%s)", string(body))
	}
}

// User controller
type ControllerUser struct{}

func (this *ControllerUser) table() string {
	return "__user"
}

func (this *ControllerUser) Path() string {
	return "user"
}

func (this *ControllerUser) Id() (string, string) {
	return "id", "[0-9]+"
}

func (this *ControllerUser) Model() grest.Model {
	id := grest.INT64("id", usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	login := grest.FIELD("login", func(value string) (interface{}, error) { return value, nil }, usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	login.SetValidate(func(i interface{}) bool {
		if s, ok := i.(string); ok && len(s) > 0 {
			return regexp.MustCompile(`^[A-Za-z_-]+$`).MatchString(s)
		}
		return false
	})
	password := grest.TEXT("password", usr.P_WO(RoleAdmin))
	name := grest.TEXT("name", usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	session := grest.EXPAND("session", []grest.Field{id}, UserSession.Model(), []grest.Field{grest.INT8("user_id")}, -1, RoleUser, RoleAdmin)
	return grest.NewModel(this.table(), []grest.Field{id, login, password, name}, session)
}

func (this *ControllerUser) Actions() []grest.Action {
	token := grest.NewAction(grest.MethodPost, "token", func(_ *grest.Request) (i int, m map[string]string, i2 interface{}, err error) {
		return http.StatusOK, nil, "user.Token", nil
	}, RoleUser)
	return []grest.Action{
		token,
		grest.NewActionList(RoleUser, RoleAdmin),
		grest.NewActionView(RoleUser, RoleAdmin),
		grest.NewActionCreate(RoleAdmin),
		grest.NewActionUpdate(RoleAdmin),
		grest.NewActionDelete(RoleAdmin),
	}
}

func (this *ControllerUser) Migrations() []db.Migration {
	return []db.Migration{
		db.NewMigration("v-user-0001",
			fmt.Sprintf(`CREATE TABLE "%s" (
                            "id" INTEGER PRIMARY KEY AUTOINCREMENT,
                            "login" TEXT NOT NULL,
                            "password" TEXT NOT NULL,
                            "name" TEXT
                          );`, this.table()),
			fmt.Sprintf(`DROP TABLE "%s";`, this.table()),
		),
		db.NewMigration("v-user-0002",
			fmt.Sprintf(`CREATE UNIQUE INDEX "uidx_%s_login" ON %s ("login");`, this.table(), this.table()),
			fmt.Sprintf(`DROP INDEX "uidx_%s_login";`, this.table()),
		),
		db.NewMigration("v-user-0003",
			fmt.Sprintf(`INSERT INTO %s ("login", "password", "name")
                          VALUES ('admin','pass-admin','ADMIN'),
                                 ('support','pass-support','SUPPORT'),
                                 ('user','pass-user','USER');`, this.table()),
			fmt.Sprintf(`DELETE FROM "%s";`, this.table()),
		),
	}
}

// User session controller
type ControllerUserSession struct{}

func (this *ControllerUserSession) table() string {
	return "__user_session"
}

func (this *ControllerUserSession) Path() string {
	return "user-session"
}

func (this *ControllerUserSession) Id() (string, string) {
	return "created_at", ""
}

func (this *ControllerUserSession) Model() grest.Model {
	userId := grest.INT8("user_id", usr.P_RO(RoleAdmin))
	ip := grest.TEXT("ip", usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	system := grest.TEXT("os", usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	createdAt := grest.INT32("created_at", usr.P_RW(RoleAdmin))
	operatingSystem := grest.EXPAND("operating-system", []grest.Field{system}, OperatingSystem.Model(), []grest.Field{grest.TEXT("name")}, 1, RoleUser, RoleAdmin)
	return grest.NewModel(this.table(), []grest.Field{userId, ip, system, createdAt}, operatingSystem)
}

func (this *ControllerUserSession) Actions() []grest.Action {
	return []grest.Action{
		grest.NewActionPagination(),
		grest.NewActionView(),
	}
}

func (this *ControllerUserSession) Migrations() []db.Migration {
	return []db.Migration{
		db.NewMigration("v-user_session-0001",
			fmt.Sprintf(`CREATE TABLE "%s" (
                            "user_id" INTEGER NOT NULL,
                            "ip" TEXT NOT NULL,
                            "os" TEXT NOT NULL DEFAULT '',
                            "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           PRIMARY KEY ("user_id", "created_at")
                          );`, this.table()),
			fmt.Sprintf(`DROP TABLE "%s";`, this.table()),
		),
		db.NewMigration("v-user_session-0002",
			fmt.Sprintf(`INSERT INTO %s ("user_id", "ip", "os", "created_at")
                          VALUES (3,'3.0.0.001','LINUX', 3001),
                                 (3,'3.0.0.002','LINUX', 3002),
                                 (3,'3.0.0.003','LINUX', 3003),
                                 (3,'3.0.0.004','LINUX', 3004),
                                 (3,'3.0.0.005','LINUX', 3005),
                                 (2,'2.0.0.001','LINUX', 2001),
                                 (2,'2.0.0.002','LINUX', 2002),
                                 (2,'2.0.0.003','LINUX', 2003),
                                 (2,'2.0.0.004','LINUX', 2004),
                                 (2,'2.0.0.005','LINUX', 2005),
                                 (2,'2.0.0.006','LINUX', 2006),
                                 (2,'2.0.0.007','LINUX', 2007),
                                 (1,'1.0.0.001','LINUX', 1001),
                                 (1,'1.0.0.002','LINUX', 1002),
                                 (1,'1.0.0.003','LINUX', 1003),
                                 (1,'1.0.0.004','LINUX', 1004),
                                 (1,'1.0.0.005','LINUX', 1005),
                                 (1,'1.0.0.006','LINUX', 1006),
                                 (1,'1.0.0.007','LINUX', 1007),
                                 (1,'1.0.0.008','LINUX', 1008),
                                 (1,'1.0.0.009','LINUX', 1009),
                                 (1,'1.0.0.010','LINUX', 1010),
                                 (1,'1.0.0.011','LINUX', 1011),
                                 (1,'1.0.0.012','LINUX', 1012),
                                 (1,'1.0.0.013','LINUX', 1013),
                                 (1,'1.0.0.014','LINUX', 1014),
                                 (1,'1.0.0.015','LINUX', 1015),
                                 (1,'1.0.0.016','LINUX', 1016),
                                 (1,'1.0.0.017','LINUX', 1017),
                                 (1,'1.0.0.018','LINUX', 1018),
                                 (1,'1.0.0.019','LINUX', 1019),
                                 (1,'1.0.0.020','LINUX', 1020),
                                 (1,'1.0.0.021','LINUX', 1021),
                                 (1,'1.0.0.022','LINUX', 1022),
                                 (1,'1.0.0.023','LINUX', 1023),
                                 (1,'1.0.0.024','LINUX', 1024);`, this.table()),
			fmt.Sprintf(`DELETE FROM "%s";`, this.table()),
		),
	}
}

// Operating system controller
type ControllerOperatingSystem struct{}

func (this *ControllerOperatingSystem) table() string {
	return "__operating_system"
}

func (this *ControllerOperatingSystem) Path() string {
	return "operating-system"
}

func (this *ControllerOperatingSystem) Id() (string, string) {
	return "name", "\\w+"
}

func (this *ControllerOperatingSystem) Model() grest.Model {
	name := grest.TEXT("name", usr.P_RO(RoleUser), usr.P_RO(RoleAdmin))
	description := grest.TEXT("description", usr.P_RO(RoleUser), usr.P_RW(RoleAdmin))
	return grest.NewModel(this.table(), []grest.Field{name, description})
}

func (this *ControllerOperatingSystem) Actions() []grest.Action {
	return []grest.Action{
		grest.NewActionList(),
		grest.NewActionView(),
	}
}

func (this *ControllerOperatingSystem) Migrations() []db.Migration {
	return []db.Migration{
		db.NewMigration("v-operating_system-0001",
			fmt.Sprintf(`CREATE TABLE "%s" (
                            "name" TEXT NOT NULL,
                            "description" TEXT NOT NULL,
                           PRIMARY KEY ("name")
                          );
                          INSERT INTO %s ("name", "description")
                          VALUES ('LINUX', 'LINUX-x86'),
                                 ('LINUX64', 'LINUX-x86-64'),
                                 ('WINDOWS', 'WINDOWS-x86'),
                                 ('WINDOWS64', 'WINDOWS-x64');`, this.table(), this.table()),
			fmt.Sprintf(`DELETE FROM %s;
                          DROP TABLE "%s";`, this.table(), this.table()),
		),
	}
}

// Operating system controller
type ControllerAPIDocs struct{}

func (this *ControllerAPIDocs) Path() string {
	return "api/docs"
}

func (this *ControllerAPIDocs) Actions() []grest.Action {
	controllers := &grest.ModuleControllersShare{CSS: []string{}, Role: []usr.Role{RoleAdmin}}
	controllers.SetPath("controller")
	request := &grest.ModuleHttpParser{Role: []usr.Role{RoleAdmin, RoleUser}}
	request.SetPath("request/parser")
	query := &grest.ModuleSqlEditor{Role: []usr.Role{RoleAdmin, RoleUser}}
	query.SetPath("sql/editor")
	return []grest.Action{request, query, controllers}
}

// http query
func httpQuery(method, url string, head map[string]string, body map[string]interface{}) (int, map[string][]string, []byte, error) {
  // create request
  if req, err := http.NewRequest(method, url, nil); err != nil {
    return 0, nil, nil, err
  } else if req != nil {
    // request header
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Cache-Control", "no-cache")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Pragma", "no-cache")
    if head != nil && len(head) > 0 {
      for k, v := range head {
        req.Header.Set(k, v)
      }
    }
    // request data
    if body != nil && len(body) > 0 {
      data, err := json.Marshal(body)
      if err != nil {
        return 0, nil, nil, err
      }
      req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
    }
    // send
    clt := &http.Client{}
    if resp, err := clt.Do(req); err != nil {
      return 0, nil, nil, err
    } else if resp != nil {
      defer func() {
        _ = resp.Body.Close()
      }()
      if b, err := ioutil.ReadAll(resp.Body); err != nil {
        return resp.StatusCode, resp.Header, nil, err
      } else {
        return resp.StatusCode, resp.Header, b, err
      }
    }
  }
  return 0, nil, nil, fmt.Errorf("missing request body")
}
