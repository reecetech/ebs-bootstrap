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
