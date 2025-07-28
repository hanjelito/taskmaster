package process

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// UserCredentials contiene UID y GID para un usuario
type UserCredentials struct {
	UID uint32
	GID uint32
}

// LookupUser busca un usuario y grupo en el sistema
func LookupUser(username, groupname string) (*UserCredentials, error) {
	// Si no se especifica usuario, no hacer de-escalation
	if username == "" {
		return nil, nil
	}

	// Verificar que somos root
	if os.Geteuid() != 0 {
		return nil, fmt.Errorf("privilege de-escalation requires running as root (current UID: %d)", os.Geteuid())
	}

	// Buscar usuario
	u, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("user '%s' not found: %v", username, err)
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid UID for user '%s': %v", username, err)
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid GID for user '%s': %v", username, err)
	}

	// Si se especifica grupo, usar ese en lugar del grupo primario del usuario
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return nil, fmt.Errorf("group '%s' not found: %v", groupname, err)
		}

		groupGid, err := strconv.ParseUint(g.Gid, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid GID for group '%s': %v", groupname, err)
		}
		gid = groupGid
	}

	return &UserCredentials{
		UID: uint32(uid),
		GID: uint32(gid),
	}, nil
}

// ApplyCredentials aplica las credenciales a SysProcAttr
func ApplyCredentials(attr *syscall.SysProcAttr, creds *UserCredentials) {
	if creds == nil {
		return
	}

	attr.Credential = &syscall.Credential{
		Uid: creds.UID,
		Gid: creds.GID,
	}
}