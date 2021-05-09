package main

import (
  "fmt"
  "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

  . "jackc_pgx-example/driver"

  "github.com/jackc/pgx"
  "github.com/prorochestvo/grest"
  "github.com/prorochestvo/grest/db"
  "github.com/prorochestvo/grest/usr"
)

const (
	PSQLHost string = "127.0.0.1"
	PSQLPort uint16 = 5432
	PSQLBase string = "test"
	PSQLUser string = "login"
	PSQLPass string = "password"

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

func (this *users) Path() string { return this.table() }

func (this *users) Id() (string, string) { return "id", "[0-9]+" }

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
			fmt.Sprintf(`
CREATE TABLE "%s" (
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
			fmt.Sprintf(`
INSERT INTO "%s" ("login", "password", "name")
VALUES ('admin','pass-admin','ADMIN'),
       ('support','pass-support','SUPPORT'),
       ('user','pass-user','USER');`, this.table()),
			fmt.Sprintf(`DELETE FROM "%s";`, this.table()),
		),
	}
}

func main() {
	// config
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
	port := HTTPPort
	if i, err := strconv.ParseUint(os.Getenv("HTTP_PORT"), 10, 16); err == nil {
		port = uint16(i)
	}
	// database connection
  dbase, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: psqlConfig})
	if err != nil {
		log.Fatalf("PSQL %s", err.Error())
	}
	// rest router
	router := grest.NewJSONRouter(Driver(dbase))
  router.Version = "1.0"
	router.AccessControl.User = func(r *grest.Request) (usr.User, error) {
		role, _ := strconv.ParseInt(r.Header.Get("Authorization"), 10, 32)
		return usr.NewUser(r.Header.Get("Authorization"), usr.Role(role)), nil
	}
	if err := router.Listen(&users{}); err != nil {
		log.Fatalf("GREST CONTROLLER: %s", err.Error())
	}
	// migrate db
	if err := router.Migration.Up(); err != nil {
		log.Fatalf("GREST MIGRATION: %s", err.Error())
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
	srv.Addr = fmt.Sprintf(":%d", port)
	srv.WriteTimeout = 30 * time.Second
	srv.ReadTimeout = 30 * time.Second
	log.Printf("listen: 127.0.0.1:%d\n", port)
	if err := srv.ListenAndServe(); err != nil {
    log.Fatalf("HTTP SERVER: %s", err.Error())
	}
}
