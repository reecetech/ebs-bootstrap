package model

import (
	"fmt"
	"io/fs"
	"strconv"
)

type FileType uint32

const (
	RegularFile FileType = 1
	Directory   FileType = 2
	Special     FileType = 3
)

type File struct {
	Path     string
	Type     FileType
	DeviceId uint64
	InodeNo  uint64
	UserId
	GroupId
	Permissions FilePermissions
}

type FilePermissions uint32

// It is useful to be able to convert FilePermissions back into the fs.FileMode
// type which is expected by Go standard libraries
func (p FilePermissions) Perm() fs.FileMode {
	return fs.FileMode(p)
}

// Linux File Permission bits are typically represented as octals: e.g 0755.
// Some users may feel comfortable representing file permission bits as decimals:
// e.g 755. While the latter is not considered an octal, lets not punish them
// for a behaviour that has been ingrained by tools like chmod.
// `strconv.ParseUint` has the ability to force the intepreation of a string as a base-8
// unsigned integer
func (p *FilePermissions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ps string
	if err := unmarshal(&ps); err != nil {
		return err
	}
	if len(ps) == 0 {
		*p = FilePermissions(0)
		return nil
	}
	// Base: 8, Bit Length: 32
	mode, err := strconv.ParseUint(ps, 8, 32)
	if err != nil {
		return fmt.Errorf("🔴 invalid permission value. '%v' must be a valid octal number", ps)
	}
	if mode > 0777 {
		return fmt.Errorf("🔴 invalid permission value. '%#o' exceeds the maximum allowed value (0777)", mode)
	}
	*p = FilePermissions(mode)
	return nil
}
