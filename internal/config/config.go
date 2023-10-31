package config

import (
	"os"
	"fmt"
	"flag"
	"bytes"
    "gopkg.in/yaml.v2"
	"ebs-bootstrap/internal/service"
)

type ConfigDevice struct {
	Fs     		string `yaml:"fs"`
	MountPoint  string `yaml:"mount_point"`
	Owner  		string `yaml:"owner"`
	Group  		string `yaml:"group"`
	Label  		string `yaml:"label"`
	Permissions string `yaml:"permissions"`
	Mode		string `yaml:"mode"`
}

type ConfigGlobal struct {
	Mode		string `yaml:"mode"`
}

type Config struct {
	Global 		ConfigGlobal 				`yaml:"global"`
	Devices 	map[string]ConfigDevice		`yaml:"devices"`
}

func New(args []string, dt *service.DeviceTranslator, fs service.FileService) (*Config, error) {
	// Generate path of config
	cp, err := parseFlags(args[0], args[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, err)
        return nil, fmt.Errorf("🔴 Failed to parse provided flags")
    }

    // Validate the path first
    if err := fs.ValidateFile(cp); err != nil {
        return nil, err
    }

    // Create config structure
    config := &Config{}

    // Load config file into memory as byte[]
    file, err := os.ReadFile(cp)
    if err != nil {
		return nil, err
    }

    // Unmarshal YAML file from memory into struct
	err = yaml.UnmarshalStrict(file, config)
    if err != nil {
		fmt.Fprintln(os.Stderr, err)
        return nil, fmt.Errorf("🔴 Failed to ingest malformed config")
    }

	// Layer modifications to the Config. These modifiers will incrementally
	// transform the Config until it reaches a desired state
	modifiers := []Modifiers{
		&OwnerModifier{},
		&DeviceModifier{
			DeviceTranslator: dt,
		},
		&GroupModifier{},
		&DeviceModeModifier{},
	}
	for _, modifier := range modifiers {
		err = modifier.Modify(config)
		if err != nil {
			return nil, err
		}
	}
    return config, nil
}

func parseFlags(program string, args []string) (string, error) {
	flags := flag.NewFlagSet(program, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

    // String that contains the configured configuration path
    var cp string
	// String that contains the mode of bootstrap process

    // Set up a CLI flag called "-config" to allow users
    // to supply the configuration file
    flags.StringVar(&cp, "config", "/etc/ebs-bootstrap/config.yml", "path to config file")

    // Actually parse the flag
    err := flags.Parse(args)
	if err != nil {
		return "", fmt.Errorf(buf.String())
	}
    // Return the configuration path
    return cp, nil
}
