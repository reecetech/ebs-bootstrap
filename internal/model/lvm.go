package model

import "github.com/reecetech/ebs-bootstrap/internal/datastructures"

type PhysicalVolume struct {
	Name string
}

type VolumeGroup struct {
	Name           string
	PhysicalVolume string
}

type LogicalVolumeState int32

const (
	Inactive    LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeInactive)
	Active      LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeActive)
	Unsupported LogicalVolumeState = LogicalVolumeState(datastructures.LogicalVolumeUnsupported)
)

type LogicalVolume struct {
	Name        string
	VolumeGroup string
	State       LogicalVolumeState
}
