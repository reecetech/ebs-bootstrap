package layer

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestCreateVolumeGroupLayerShouldProcess(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		ExpectedValue bool
	}{
		{
			Name: "At Least Once Device Has Lvm Specified",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						Lvm: "lvm-id",
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedValue: true,
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
			cvgl := NewCreateVolumeGroupLayer(nil)
			output := cvgl.ShouldProcess(subtest.Config)
			utils.CheckOutput("cvgl.ShouldProcess()", t, subtest.ExpectedValue, output)
		})
	}
}
