package usr

var (
	DefaultRole Role = 0
)

type Role uint16

type Roles []Role

func (this Roles) IndexOf(item Role) int {
	for i, r := range this {
		if r == item {
			return i
		}
	}
	return -1
}
