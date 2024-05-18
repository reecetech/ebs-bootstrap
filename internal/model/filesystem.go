package model

import "fmt"

type FileSystem string

const (
	Unformatted FileSystem = ""
	Ext4        FileSystem = "ext4"
	Xfs         FileSystem = "xfs"
	Lvm         FileSystem = "LVM2_member"
)

func (fs FileSystem) String() string {
	if len(fs) == 0 {
		return "unformatted"
	}
	return string(fs)
}

func ParseFileSystem(s string) (FileSystem, error) {
	fst := FileSystem(s)
	switch fst {
	case Unformatted, Ext4, Xfs, Lvm:
		return fst, nil
	default:
		return fst, fmt.Errorf("File system '%s' is not supported", fst.String())
	}
}
