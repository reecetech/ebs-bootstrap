package config

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"gopkg.in/yaml.v2"
)

const (
	DefaultMode           = model.Healthcheck
	DefaultMountOptions   = model.MountOptions("defaults")
	DefaultLvmConsumption = 100
)

type Flag struct {
	Config         string
	Mode           string
	Remount        bool
	MountOptions   string
	Resize         bool
	LvmConsumption int
}

type Device struct {
	Fs          model.FileSystem      `yaml:"fs"`
	MountPoint  string                `yaml:"mountPoint"`
	User        string                `yaml:"user"`
	Group       string                `yaml:"group"`
	Label       string                `yaml:"label"`
	Permissions model.FilePermissions `yaml:"permissions"`
	Lvm         string                `yaml:"lvm"`
	Options     `yaml:",inline"`
}

type Options struct {
	Mode           model.Mode         `yaml:"mode"`
	Remount        bool               `yaml:"remount"`
	MountOptions   model.MountOptions `yaml:"mountOptions"`
	Resize         bool               `yaml:"resize"`
	LvmConsumption int                `yaml:"lvmConsumption"`
}

// We don't export "overrides" as this is an attribute that is used
// internally to store the state of flag overrides
type Config struct {
	Defaults  Options           `yaml:"defaults"`
	Devices   map[string]Device `yaml:"devices"`
	overrides Options
}

func New(args []string) (*Config, error) {
	// Generate config path
	f, err := parseFlags(args[0], args[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return nil, fmt.Errorf("🔴 Failed to parse provided flags")
	}

	// Load config file into memory
	file, err := os.ReadFile(f.Config)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("🔴 %s: File not found", f.Config)
		}
		return nil, fmt.Errorf("🔴 %s: %v", f.Config, err)
	}

	// Create config structure
	c := &Config{}

	// Unmarshal YAML file from memory into struct
	err = yaml.UnmarshalStrict(file, c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, fmt.Errorf("🔴 %s: Failed to ingest malformed config", f.Config)
	}

	// Inject flag overrides into config
	return c.setOverrides(f), nil
}

func parseFlags(program string, args []string) (*Flag, error) {
	flags := flag.NewFlagSet(program, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

	// String that contains the configured configuration path
	f := &Flag{}
	// String that contains the mode of bootstrap process

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flags.StringVar(&f.Config, "config", "/etc/ebs-bootstrap/config.yml", "path to config file")
	flags.StringVar(&f.Mode, "mode", "", "override for mode")
	flags.BoolVar(&f.Remount, "remount", false, "override for remount")
	flags.StringVar(&f.MountOptions, "mount-options", "", "override for mount options")
	flags.BoolVar(&f.Resize, "resize", false, "override for resize filesystem")
	flags.IntVar(&f.LvmConsumption, "lvm-consumption", 0, "override for lvm consumption")

	// Actually parse the flag
	err := flags.Parse(args)
	if err != nil {
		return nil, fmt.Errorf(buf.String())
	}

	return f, nil
}

func (c *Config) setOverrides(f *Flag) *Config {
	c.overrides.Mode = model.Mode(f.Mode)
	c.overrides.Remount = f.Remount
	c.overrides.MountOptions = model.MountOptions(f.MountOptions)
	c.overrides.Resize = f.Resize
	c.overrides.LvmConsumption = f.LvmConsumption
	return c
}

func (c *Config) GetMode(name string) model.Mode {
	cd, found := c.Devices[name]
	if !found {
		return DefaultMode
	}
	if c.overrides.Mode != model.Empty {
		return c.overrides.Mode
	}
	if cd.Mode != model.Empty {
		return cd.Mode
	}
	if c.Defaults.Mode != model.Empty {
		return c.Defaults.Mode
	}
	return DefaultMode
}

func (c *Config) GetRemount(name string) bool {
	cd, found := c.Devices[name]
	if !found {
		return false
	}
	return c.overrides.Remount || c.Defaults.Remount || cd.Remount
}

func (c *Config) GetMountOptions(name string) model.MountOptions {
	cd, found := c.Devices[name]
	if !found {
		return DefaultMountOptions
	}
	if len(c.overrides.MountOptions) > 0 {
		return c.overrides.MountOptions
	}
	if len(cd.MountOptions) > 0 {
		return cd.MountOptions
	}
	if len(c.Defaults.MountOptions) > 0 {
		return c.Defaults.MountOptions
	}
	return DefaultMountOptions
}

func (c *Config) GetResize(name string) bool {
	cd, found := c.Devices[name]
	if !found {
		return false
	}
	return c.overrides.Resize || c.Defaults.Resize || cd.Resize
}

func (c *Config) GetLvmConsumption(name string) int {
	cd, found := c.Devices[name]
	if !found {
		return DefaultLvmConsumption
	}
	if c.overrides.LvmConsumption > 0 {
		return c.overrides.LvmConsumption
	}
	if cd.LvmConsumption > 0 {
		return cd.LvmConsumption
	}
	if c.Defaults.LvmConsumption > 0 {
		return c.Defaults.LvmConsumption
	}
	return DefaultLvmConsumption
}
