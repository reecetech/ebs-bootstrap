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
	DefaultMode         = model.Healthcheck
	DefaultMountOptions = model.MountOptions("defaults")
)

type Flag struct {
	Config          string
	Mode            string
	Remount         bool
	MountOptions    string
	ResizeFs        bool
	ResizeThreshold float64
}

type Device struct {
	Fs          model.FileSystem      `yaml:"fs"`
	MountPoint  string                `yaml:"mountPoint"`
	User        string                `yaml:"user"`
	Group       string                `yaml:"group"`
	Label       string                `yaml:"label"`
	Permissions model.FilePermissions `yaml:"permissions"`
	Options     `yaml:",inline"`
}

type Options struct {
	Mode            model.Mode         `yaml:"mode"`
	Remount         bool               `yaml:"remount"`
	MountOptions    model.MountOptions `yaml:"mountOptions"`
	ResizeFs        bool               `yaml:"resizeFs"`
	ResizeThreshold float64            `yaml:"resizeThreshold"`
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
	flags.BoolVar(&f.ResizeFs, "resize-fs", false, "override for resize filesystem")
	flags.Float64Var(&f.ResizeThreshold, "resize-threshold", 0, "override for resize threshold")

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
	c.overrides.ResizeFs = f.ResizeFs
	c.overrides.ResizeThreshold = f.ResizeThreshold
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

func (c *Config) GetResizeFs(name string) bool {
	cd, found := c.Devices[name]
	if !found {
		return false
	}
	return c.overrides.ResizeFs || c.Defaults.ResizeFs || cd.ResizeFs
}

func (c *Config) GetResizeThreshold(name string) float64 {
	cd, found := c.Devices[name]
	if !found {
		return 0
	}
	if c.overrides.ResizeThreshold > 0 {
		return c.overrides.ResizeThreshold
	}
	if cd.ResizeThreshold > 0 {
		return cd.ResizeThreshold
	}
	return c.Defaults.ResizeThreshold
}
