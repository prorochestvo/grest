package usr

var (
	DefaultUser User = NewUser(nil, DefaultRole)
)

type User interface {
	ID() interface{}
	Role() Role
}

func NewUser(id interface{}, role Role) User {
	result := user{}
	result.id = id
	result.role = role
	return &result
}

type user struct {
	id   interface{}
	role Role
}

func (this *user) ID() interface{} {
	return this.id
}

func (this *user) Role() Role {
	return this.role
}
