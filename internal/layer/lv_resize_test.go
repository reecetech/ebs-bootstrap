package layer

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestResizeLogicalVolumeLayerShouldProcess(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		ExpectedValue bool
	}{
		{
			Name: "At Least Once Device Has Lvm Specified and Resize Enabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						Lvm: "lvm-id",
						Options: config.Options{
							Resize: true,
						},
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedValue: true,
		},
		{
			Name: "Device Has Lvm Specified, but Resize Disabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						Lvm: "lvm-id",
					},
				},
			},
			ExpectedValue: false,
		},
		{
			Name: "No Device Has Lvm Specified",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedValue: false,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			rlvl := NewResizeLogicalVolumeLayer(nil)
			output := rlvl.ShouldProcess(subtest.Config)
			utils.CheckOutput("rlvl.ShouldProcess()", t, subtest.ExpectedValue, output)
		})
	}
}
