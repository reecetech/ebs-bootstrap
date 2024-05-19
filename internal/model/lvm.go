package model

type PhysicalVolume struct {
	Name string
}

type VolumeGroup struct {
	Name           string
	PhysicalVolume string
}

type LogicalVolume struct {
	Name        string
	VolumeGroup string
}
