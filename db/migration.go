package db

type Migration interface {
	Version() string
	Up() string
	Down() string
}

type MigrationController interface {
	Migrations() []Migration
}

func NewMigration(version, up, down string) Migration {
	result := migration{}
	result.version = version
	result.up = up
	result.down = down
	return &result
}

type migration struct {
	version string
	up      string
	down    string
}

func (this *migration) Version() string {
	return this.version
}

func (this *migration) Up() string {
	return this.up
}

func (this *migration) Down() string {
	return this.down
}
