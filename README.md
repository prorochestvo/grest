### **GREST** *(GO-REST)*

[![Build Status](https://travis-ci.org/prorochestvo/grest.svg?branch=master)](https://travis-ci.org/prorochestvo/grest)

This small framework to quickly build RESTful server for GoLang.



### Features
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



### Installation

This package can be installed with the go get command:

`go get github.com/prorochestvo/grest`



### How to use
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

// user controller
type users struct {
}

// db table name
func (this *users) table() string {
  return "users"
}

// url path
func (this *users) Path() string {
  return this.table()
}

// db table primary key
func (this *users) Id() (string, string) {
  return "id", "[0-9]+"
}

// db table fields
func (this *users) Model() grest.Model {
  id := grest.INT64("id", usr.P_RO(RoleUser), usr.P_RO(RoleSupport), usr.P_RO(RoleAdmin))
  login := grest.TEXT("login", nil, usr.P_RO(RoleUser), usr.P_RO(RoleSupport), usr.P_RW(RoleAdmin))
  password := grest.TEXT("password", nil, usr.P_WO(RoleSupport), usr.P_WO(RoleAdmin))
  name := grest.TEXT("name", usr.P_RW(RoleUser), usr.P_RW(RoleSupport), usr.P_RW(RoleAdmin))
  return grest.NewModel(this.table(), []grest.Field{id, login, password, name})
}

// available actions
func (this *users) Actions() []grest.Action {
  return []grest.Action{
    grest.NewActionPagination(RoleUser, RoleSupport, RoleAdmin),
    grest.NewActionView(RoleUser, RoleSupport, RoleAdmin),
    grest.NewActionCreate(RoleSupport, RoleAdmin),
    grest.NewActionUpdate(RoleSupport, RoleAdmin),
    grest.NewActionDelete(RoleSupport, RoleAdmin),
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



### Database schema migration

Implement controller `Migrations() []db.Migration` method with simple sql query
###### Example:
```
// controller version steps
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
```

Migration steps manager:

Operation | Method | Description
--- | --- | ---
 UP  | `Migration.Up() error` | execute all migration steps
 DOWN | `Migration.Down() error` | rollback one migration step
 RESET | `Migration.Down() error` | rollback all migration steps
 REVISE |  `Migration.revise() error` | find the difference between model/fields and database
###### Example:
```

...

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

...

```



### URL Query parameters
> Working only pagination and list action

URL | SQL
--- | ---
:between[*field*][]=*val1*&:between[*field*][]=*val2* | **WHERE** *field* **BETWEEN** val1 **AND** val2
:is_null[*field*]                                     | **WHERE** **IS NULL** *field*
:like[*field*]=*val*                                  | **WHERE** *field* **LIKE** *val*
:in[*field*][]=*val*                                  | **WHERE** *field* **IN** (*val*)
:cmp_be[*field*]=*val*                                | **WHERE** *field* **>=** *val*
:cmp_b[*field*]=*val*                                 | **WHERE** *field* **>** *val*
:cmp_l[*field*]=*val*                                 | **WHERE** *field* **<** *val*
:cmp_le[*field*]=*val*                                | **WHERE** *field* **<=** *val*
*field*=*val*                                         | **WHERE** *field* **=** *val*
:group[*field*]                                       | **GROUP BY** *field*
:group[*field*]=VAL                                   | **GROUP BY** *field* **HAVING** *field* **=** *val*
:sort[*field*]                                        | **ORDER BY** *field* **ASC**
:sort[*field*]=DESC                                   | **ORDER BY** *field* **DESC**
:offset=*num*                                         | **LIMIT** *num*
:limit=*num*                                          | **OFFSET** *num*

##### URL Query operators:
URL | SQL | *Description*
--- | --- | -----------
:*query*       | **AND** (*query*) | *default, if not set pipe mark ( &#124; )*
:&#124;*query* | **OR** (*query*)  | *set pipe mark ( &#124; )*
:!*query*      | **NOT** (*query*) | *set exclamation mark ( &#124; )*

##### URL Query sorting:
```
  NUM:query   (NUM - position in the sql query) 
```
##### URL Query examples:
```
  http://127.0.0.1:80/user?:!in[id][]=1&:!in[id][]=2&:!in[id][]=3
  
  < SELECT *
  < FROM user
  < WHERE (NOT(id IN ('1', '2', '3')));
```
```
  http://127.0.0.1:80/user?2:|like[name]=A%25&1:like[name]=B%25
  
  < SELECT *
  < FROM user
  < WHERE (name LIKE 'B%') OR (name LIKE 'A%');
```
```
  http://127.0.0.1:80/user?:sort[name]=ASC&role=admin
  
  < SELECT *
  < FROM user
  < WHERE (role = 'admin')
  < ORDER BY name ASC;
```



### Dependences
  * golang v 1.13 or late



### External packages
  * [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
  * [jackc/pgx](https://github.com/jackc/pgx)



### Donate
  
You can support project making a [donation](https://www.paypal.me/prorochestvo). Thanks!