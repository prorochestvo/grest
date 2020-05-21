### GREST (GO REST) ###
##### Implementation of the http rest full api protocol
##### v 1.0


### Features ###
 - [x] Full REST API
 - [x] Fast controllers
 - [x] Content-Type JSON / XML
 - [x] CORS request  
 - [ ] Logging
 - [x] Pagination
 - [x] Expand model field
 - [ ] API documentation
 - [x] Migration database
 - [x] DataBase (PostgreSQL, ~~MySQL~~)


### Installation ###

This package can be installed with the go get command:

`go get github.com/prorochestvo/grest`

### How used ###
```
package main

import (
  "fmt"
  "github.com/jackc/pgx"
  "grest"
  "grest/db"
  "grest/example/postgresql"
  "grest/usr"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"
)

const (
  HTTPPort uint16 = 8080

  RoleAdmin   usr.Role = 1
  RoleSupport usr.Role = 2
  RoleUser    usr.Role = 3
)

type users struct {
}

func (this *users) table() string {
  return "users"
}

func (this *users) Path() string {
  return this.table()
}

func (this *users) Id() (string, string) {
  return "id", "[0-9]+"
}

func (this *users) Model() grest.Model {
  id := grest.INT64("id", usr.P_RO(RoleUser), usr.P_RO(RoleSupport), usr.P_RO(RoleAdmin))
  login := grest.TEXT("login", nil, usr.P_RO(RoleUser), usr.P_RO(RoleSupport), usr.P_RW(RoleAdmin))
  password := grest.TEXT("password", nil, usr.P_WO(RoleSupport), usr.P_WO(RoleAdmin))
  name := grest.TEXT("name", usr.P_RW(RoleUser), usr.P_RW(RoleSupport), usr.P_RW(RoleAdmin))
  return grest.NewModel(this.table(), []grest.Field{id, login, password, name})
}

func (this *users) Actions() []grest.Action {
  return []grest.Action{
    grest.NewActionPagination(RoleUser, RoleSupport, RoleAdmin),
    grest.NewActionView(RoleUser, RoleSupport, RoleAdmin),
    grest.NewActionCreate(RoleSupport, RoleAdmin),
    grest.NewActionUpdate(RoleSupport, RoleAdmin),
    grest.NewActionDelete(RoleSupport, RoleAdmin),
  }
}

func (this *users) Migrations() []db.Migration {
  return []db.Migration{
    db.NewMigration(
      fmt.Sprintf(`"m0001-create_table_%s`, this.table()),
      fmt.Sprintf(`CREATE TABLE "%s" (
                     "id" SERIAL NOT NULL,
                     "login" TEXT NOT NULL,
                     "password" TEXT NOT NULL,
                     "name" TEXT,
                     PRIMARY KEY ("id")
                   );`, this.table()),
      fmt.Sprintf(`DROP TABLE "%s";`, this.table()),
    ),
    db.NewMigration(
      fmt.Sprintf(`"m0002-init_table_%s`, this.table()),
      fmt.Sprintf(`INSERT INTO "%s" ("login", "password", "name")
                   VALUES ('admin','pass-admin','ADMIN'),
                          ('support','pass-support','SUPPORT'),
                          ('user','pass-user','USER');`, this.table()),
      fmt.Sprintf(`DELETE FROM "%s";`, this.table()),
    ),
  }
}

func main() {
  // database connection
  psql, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: pgx.ConnConfig{Host: "127.0.0.17", Port: 5432, User: "user", Password: "pass", Database: "test"}})
  if err != nil {
    log.Fatal(err)
  }
  // rest router
  router := grest.NewJSONRouter(postgresql.Driver(psql))
  router.AccessControl.User = func(r *grest.Request) (usr.User, error) {
    role, _ := strconv.ParseInt(r.Header.Get("Authorization"), 10, 32)
    return usr.NewUser(r.Header.Get("Authorization"), usr.Role(role)), nil
  }
  if err := router.Listen(&users{}); err != nil {
    log.Fatal(err)
  }
  // migrate db
  if err := router.Migration.Up(); err != nil {
    log.Fatal(err)
  }
  // revise db
  if err := router.Migration.Revise(); err != nil {
    log.Println("revise:")
    for _, e := range strings.Split(err.Error(), "\n") {
      log.Printf(" - %s\n", strings.Trim(e, "\n\r\t "))
    }
  }
  // http server
  srv := http.Server{}
  srv.Handler = router
  srv.Addr = fmt.Sprintf(":%d", HTTPPort)
  srv.WriteTimeout = 30 * time.Second
  srv.ReadTimeout = 30 * time.Second
  log.Printf("Listen: 127.0.0.1:%d\n", HTTPPort)
  if err := srv.ListenAndServe(); err != nil {
    log.Fatal(err)
  }
}
```

### External packages ###
  * [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
  * [jackc/pgx](https://github.com/jackc/pgx)