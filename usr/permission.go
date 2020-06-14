package usr

import "github.com/prorochestvo/grest/internal"

const (
	ALEVEL_READ  internal.AccessLevel = 0x0001
	ALEVEL_WRITE internal.AccessLevel = 0x0002
)

type Permission interface {
	Role() Role
	Access(accessLevel ...internal.AccessLevel) bool
}

func P_RO(role Role) Permission {
	return newPermission(role, ALEVEL_READ)
}

func P_WO(role Role) Permission {
	return newPermission(role, ALEVEL_WRITE)
}

func P_RW(role Role) Permission {
	return newPermission(role, ALEVEL_READ, ALEVEL_WRITE)
}

func newPermission(role Role, accessLevel ...internal.AccessLevel) Permission {
	var result uint32 = 0
	if accessLevel != nil && len(accessLevel) > 0 {
		for _, level := range accessLevel {
			result |= (uint32(level&0xFFFF) << 16) & 0xFFFF0000
		}
	}
	result |= uint32(role) & 0x0000FFFF
	return permission(result)
}

type permission uint32

func (this permission) Role() Role {
	return Role(this & 0x0000FFFF)
}

func (this permission) Access(accessLevel ...internal.AccessLevel) bool {
	if accessLevel == nil || len(accessLevel) == 0 {
		return false
	}
	var permissions = uint32(this & 0xFFFF0000)
	var value uint32 = 0
	for _, level := range accessLevel {
		value |= (uint32(level&0xFFFF) << 16) & 0xFFFF0000
	}
	return (permissions & value) == value
}
