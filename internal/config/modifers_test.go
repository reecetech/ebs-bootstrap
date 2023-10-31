package config

import (
	"testing"
	"fmt"
	"ebs-bootstrap/internal/utils"
	"ebs-bootstrap/internal/service"
	"github.com/google/go-cmp/cmp"
)

func TestOwnerModifier(t *testing.T) {
	subtests := []struct{
		Name			string
		Config			*Config
		ExpectedOutput	*Config
		ExpectedErr		error
	}{
		{
			Name: "Existing Owner (Non-Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "root",
					},
				},
			},
			ExpectedOutput: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "0",
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Existing Owner (Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "0",
					},
				},
			},
			ExpectedOutput: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "0",
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Non-existent Owner (Non-Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "non-existent-user",
					},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("user: unknown user non-existent-user"),
		},
		{
			Name: "Non-existent Owner (Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Owner: "-1",
					},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("user: unknown userid -1"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			om := &OwnerModifier{}
			err := om.Modify(subtest.Config)
			if subtest.ExpectedOutput != nil {
				if !cmp.Equal(subtest.Config, subtest.ExpectedOutput) {
					t.Errorf("Modify() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, subtest.Config)
				}
			}
			utils.CheckError("Modify()", t, subtest.ExpectedErr, err)
		})
	}
}

func TestGroupModifier(t *testing.T) {
	_, g, err := utils.GetCurrentUserGroup()
	if err != nil {
		t.Error(err)
		return
	}
	subtests := []struct{
		Name			string
		Config			*Config
		ExpectedOutput	*Config
		ExpectedErr		error
	}{
		{
			Name: "Existing Group (Non-Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: g.Name,
					},
				},
			},
			ExpectedOutput: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: g.Gid,
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Existing Group (Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: g.Gid,
					},
				},
			},
			ExpectedOutput: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: g.Gid,
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Non-existent Group (Non-Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: "non-existent-group",
					},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("group: unknown group non-existent-group"),
		},
		{
			Name: "Non-existent Group (Int)",
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Group: "-1",
					},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("group: unknown groupid -1"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			gm := &GroupModifier{}
			err := gm.Modify(subtest.Config)
			if subtest.ExpectedOutput != nil {
				if !cmp.Equal(subtest.Config, subtest.ExpectedOutput) {
					t.Errorf("Modify() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, subtest.Config)
				}
			}
			utils.CheckError("Modify()", t, subtest.ExpectedErr, err)
		})
	}
}

func TestDeviceModifier(t *testing.T) {
	subtests := []struct{
		Name				string
		DeviceTranslator	*service.DeviceTranslator
		Config				*Config
		ExpectedOutput		*Config
		ExpectedErr			error
	}{
		{
			Name: "DeviceTranslator() Hit",
			DeviceTranslator:	&service.DeviceTranslator{
				Table:	map[string]string{
					"/dev/nvme0n1": "/dev/nvme0n1",
					"/dev/xvdf": 	"/dev/nvme0n1",
				},
			},
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvdf": ConfigDevice{},
				},
			},
			ExpectedOutput: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/nvme0n1": ConfigDevice{},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "DeviceTranslator() Miss",
			DeviceTranslator:	&service.DeviceTranslator{
				Table:	map[string]string{},
			},
			Config: &Config{
				Devices: map[string]ConfigDevice{
					"/dev/xvdf": ConfigDevice{},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("🔴 Could not identify a device with an alias /dev/xvdf"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			dm := &DeviceModifier{DeviceTranslator: subtest.DeviceTranslator}
			err := dm.Modify(subtest.Config)
			if subtest.ExpectedOutput != nil {
				if !cmp.Equal(subtest.Config, subtest.ExpectedOutput) {
					t.Errorf("Modify() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, subtest.Config)
				}
			}
			utils.CheckError("Modify()", t, subtest.ExpectedErr, err)
		})
	}
}

func TestDeviceModeModifier(t *testing.T) {
	subtests := []struct{
		Name			string
		Config			*Config
		ExpectedOutput	*Config
		ExpectedErr		error
	}{
		{
			Name: "Valid Global Mode, Empty Local Mode",
			Config: &Config{
				Global: ConfigGlobal{
					Mode: ValidDeviceModes[0],
				},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{},
				},
			},
			ExpectedOutput: &Config{
				Global: ConfigGlobal{
					Mode: ValidDeviceModes[0],
				},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Mode: ValidDeviceModes[0],
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Empty Global Mode, Valid Local Mode",
			Config: &Config{
				Global: ConfigGlobal{},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Mode: ValidDeviceModes[0],
					},
				},
			},
			ExpectedOutput: &Config{
				Global: ConfigGlobal{},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Mode: ValidDeviceModes[0],
					},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Empty Global Mode, Empty Local Mode",
			Config: &Config{
				Global: ConfigGlobal{},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("🔴 /dev/xvda: If mode is not provided locally, it must be provided globally"),
		},
		{
			Name: "Invalid Global Mode, Empty Local Mode",
			Config: &Config{
				Global: ConfigGlobal{
					Mode:	"non-supported-mode",
				},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("🔴 A valid global mode was not provided: Expected=%s Provided=non-supported-mode", ValidDeviceModes),
		},
		{
			Name: "Empty Global Mode, Invalid Local Mode",
			Config: &Config{
				Global: ConfigGlobal{},
				Devices: map[string]ConfigDevice{
					"/dev/xvda": ConfigDevice{
						Mode: "non-supported-mode",
					},
				},
			},
			ExpectedOutput: nil,
			ExpectedErr: fmt.Errorf("🔴 /dev/xvda: A valid mode was not provided: Expected=%s Provided=non-supported-mode", ValidDeviceModes),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			gm := &DeviceModeModifier{}
			err := gm.Modify(subtest.Config)
			if subtest.ExpectedOutput != nil {
				if !cmp.Equal(subtest.Config, subtest.ExpectedOutput) {
					t.Errorf("Modify() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, subtest.Config)
				}
			}
			utils.CheckError("Modify()", t, subtest.ExpectedErr, err)
		})
	}
}
