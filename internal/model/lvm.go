package model

type LvmDevice struct {
	Name  string
	Child *PhysicalVolume
}

type PhysicalVolume struct {
	Name   string
	Parent *LvmDevice
	Child  *VolumeGroup
}

type VolumeGroup struct {
	Name     string
	Parent   []*PhysicalVolume
	Children []*LogicalVolume
}

type LogicalVolume struct {
	Name   string
	Parent *VolumeGroup
	Active bool
}
