package model

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestRemount(t *testing.T) {
	subtests := []struct {
		Name string
		MountOptions
		ExpectedOutput MountOptions
	}{
		{
			Name:           "Empty",
			MountOptions:   MountOptions(""),
			ExpectedOutput: MountOptions("remount"),
		},
		{
			Name:           "Remount Not Present",
			MountOptions:   MountOptions("defaults"),
			ExpectedOutput: MountOptions("defaults,remount"),
		},
		{
			Name:           "Remount Alreay Present",
			MountOptions:   MountOptions("defaults,remount"),
			ExpectedOutput: MountOptions("defaults,remount"),
		},
	}
	for _, subtest := range subtests {
		mo := subtest.MountOptions.Remount()
		utils.CheckOutput("ParseRemount()", t, subtest.ExpectedOutput, mo)
	}
}

func TestShouldResize(t *testing.T) {
	subtests := []struct {
		Name            string
		ResizeThreshold float64
		FileSystemSize  uint64
		BlockDeviceSize uint64
		ExpectedOutput  bool
	}{
		{
			Name:            "Resize=✅",
			ResizeThreshold: 95,
			FileSystemSize:  90,
			BlockDeviceSize: 100,
			ExpectedOutput:  true,
		},
		{
			Name:            "Resize=❌",
			ResizeThreshold: 85,
			FileSystemSize:  90,
			BlockDeviceSize: 100,
			ExpectedOutput:  false,
		},
		{
			Name:            "Resize=❌ <alt>",
			ResizeThreshold: 100,
			FileSystemSize:  100,
			BlockDeviceSize: 100,
			ExpectedOutput:  false,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			bdm := &BlockDeviceMetrics{
				subtest.FileSystemSize,
				subtest.BlockDeviceSize,
			}
			sr := bdm.ShouldResize(subtest.ResizeThreshold)
			utils.CheckOutput("bdm.ShouldResize()", t, subtest.ExpectedOutput, sr)
		})
	}
}
