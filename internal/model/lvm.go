package model

type PhysicalVolume struct {
	Name string
}

type VolumeGroup struct {
	Name           string
	PhysicalVolume string
}

type LogicalVolumeState int32

const (
	Inactive    LogicalVolumeState = 0b0010000
	Active      LogicalVolumeState = 0b0110000
	Unsupported LogicalVolumeState = 0b1110000
)

type LogicalVolume struct {
	Name        string
	VolumeGroup string
	State       LogicalVolumeState
}
