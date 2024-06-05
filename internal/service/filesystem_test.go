package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestFileSystemFactory(t *testing.T) {
	fssf := NewLinuxFileSystemServiceFactory(nil)

	subtests := []struct {
		Name string
		model.FileSystem
		CmpOption      cmp.Option
		ExpectedOutput FileSystemService
		ExpectedError  error
	}{
		{
			Name:           "ext4",
			FileSystem:     model.Ext4,
			CmpOption:      cmp.AllowUnexported(Ext4Service{}),
			ExpectedOutput: NewExt4Service(nil),
			ExpectedError:  nil,
		},
		{
			Name:           "xfs",
			FileSystem:     model.Xfs,
			CmpOption:      cmp.AllowUnexported(XfsService{}),
			ExpectedOutput: NewXfsService(nil),
			ExpectedError:  nil,
		},
		{
			Name:           "brtfs",
			FileSystem:     model.FileSystem("brtfs"),
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("Support for querying/modifying the 'brtfs' filesystem is lacking"),
		},
		{
			Name:           "Unformatted",
			FileSystem:     model.Unformatted,
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("An unformatted file system can not be queried/modified"),
		},
		{
			Name:           "LVM",
			FileSystem:     model.Lvm,
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("A Physical Volume cannot be queried/modified"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fss, err := fssf.Select(subtest.FileSystem)
			utils.CheckError("fssf.Select()", t, subtest.ExpectedError, err)
			utils.CheckOutput("fssf.Select()", t, subtest.ExpectedOutput, fss, subtest.CmpOption)
		})
	}
}

func TestFormat(t *testing.T) {
	subtests := []struct {
		Name              string
		Device            string
		FileSystemService func(rf utils.RunnerFactory) FileSystemService
		RunnerBinary      utils.Binary
		RunnerArgs        []string
		RunnerOutput      string
		RunnerError       error
		ExpectedError     error
	}{
		{
			Name:              "ext4",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.MkfsExt4,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.MkfsXfs,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs + Error<existing_filesystem>",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.MkfsExt4,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError: fmt.Errorf(`🔴 mkfs.xfs: /dev/xvdf appears to contain an existing filesystem (xfs).
mkfs.xfs: Use the -f option to force overwrite.`),
			ExpectedError: fmt.Errorf(`🔴 mkfs.xfs: /dev/xvdf appears to contain an existing filesystem (xfs).
mkfs.xfs: Use the -f option to force overwrite.`),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			fss := subtest.FileSystemService(mrf)
			err := fss.Format(subtest.Device)
			utils.CheckError("fss.Format()", t, subtest.ExpectedError, err)
		})
	}
}

func TestLabel(t *testing.T) {
	subtests := []struct {
		Name              string
		Device            string
		Label             string
		FileSystemService func(rf utils.RunnerFactory) FileSystemService
		RunnerBinary      utils.Binary
		RunnerArgs        []string
		RunnerOutput      string
		RunnerError       error
		ExpectedError     error
	}{
		{
			Name:              "ext4",
			Device:            "/dev/xvdf",
			Label:             "test",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.E2Label,
			RunnerArgs:        []string{"/dev/xvdf", "test"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs",
			Device:            "/dev/xvdf",
			Label:             "test",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsAdmin,
			RunnerArgs:        []string{"-L", "test", "/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs + error<permission_denied>",
			Device:            "/dev/xvdf",
			Label:             "test",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsAdmin,
			RunnerArgs:        []string{"-L", "test", "/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       fmt.Errorf("🔴 xfs_admin: cannot open /dev/vdb: Permission denied"),
			ExpectedError:     fmt.Errorf("🔴 xfs_admin: cannot open /dev/vdb: Permission denied"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			fss := subtest.FileSystemService(mrf)
			err := fss.Label(subtest.Device, subtest.Label)
			utils.CheckError("fss.Label()", t, subtest.ExpectedError, err)
		})
	}
}
func TestResize(t *testing.T) {
	subtests := []struct {
		Name              string
		Device            string
		FileSystemService func(rf utils.RunnerFactory) FileSystemService
		RunnerBinary      utils.Binary
		RunnerArgs        []string
		RunnerOutput      string
		RunnerError       error
		ExpectedError     error
	}{
		{
			Name:              "ext4",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.Resize2fs,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsGrowfs,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       nil,
			ExpectedError:     nil,
		},
		{
			Name:              "xfs + error<permission_denied>",
			Device:            "/dev/xvdf",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsGrowfs,
			RunnerArgs:        []string{"/dev/xvdf"},
			RunnerOutput:      "",
			RunnerError:       fmt.Errorf("🔴 xfs_growfs: cannot open /dev/vdb: Permission denied"),
			ExpectedError:     fmt.Errorf("🔴 xfs_growfs: cannot open /dev/vdb: Permission denied"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, subtest.RunnerOutput, subtest.RunnerError)
			fss := subtest.FileSystemService(mrf)
			err := fss.Resize(subtest.Device)
			utils.CheckError("fss.Resize()", t, subtest.ExpectedError, err)
		})
	}
}

func TestGetFileSystem(t *testing.T) {
	mrf := &utils.MockRunnerFactory{}
	subtests := []struct {
		Name string
		FileSystemService
		ExpectedOutput model.FileSystem
	}{
		{
			Name:              "ext4",
			FileSystemService: NewExt4Service(mrf),
			ExpectedOutput:    model.Ext4,
		},
		{
			Name:              "xfs",
			FileSystemService: NewXfsService(mrf),
			ExpectedOutput:    model.Xfs,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fs := subtest.FileSystemService.GetFileSystem()
			utils.CheckOutput("fss.GetFileSystem()", t, subtest.ExpectedOutput, fs)
		})
	}
}

func TestGetMaximumLabelLength(t *testing.T) {
	mrf := &utils.MockRunnerFactory{}
	subtests := []struct {
		Name string
		FileSystemService
		ExpectedOutput int
	}{
		{
			Name:              "ext4",
			FileSystemService: NewExt4Service(mrf),
			ExpectedOutput:    16,
		},
		{
			Name:              "xfs",
			FileSystemService: NewXfsService(mrf),
			ExpectedOutput:    12,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fs := subtest.FileSystemService.GetMaximumLabelLength()
			utils.CheckOutput("fss.GetMaximumLabelLength()", t, subtest.ExpectedOutput, fs)
		})
	}
}

func TestResizeRequireMount(t *testing.T) {
	mrf := &utils.MockRunnerFactory{}
	subtests := []struct {
		Name string
		FileSystemService
		ExpectedOutput bool
	}{
		{
			Name:              "ext4",
			FileSystemService: NewExt4Service(mrf),
			ExpectedOutput:    false,
		},
		{
			Name:              "xfs",
			FileSystemService: NewXfsService(mrf),
			ExpectedOutput:    true,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fs := subtest.FileSystemService.DoesResizeRequireMount()
			utils.CheckOutput("fss.DoesResizeRequireMount()", t, subtest.ExpectedOutput, fs)
		})
	}
}

func TestLabelRequireUnmount(t *testing.T) {
	mrf := &utils.MockRunnerFactory{}
	subtests := []struct {
		Name string
		FileSystemService
		ExpectedOutput bool
	}{
		{
			Name:              "ext4",
			FileSystemService: NewExt4Service(mrf),
			ExpectedOutput:    false,
		},
		{
			Name:              "xfs",
			FileSystemService: NewXfsService(mrf),
			ExpectedOutput:    true,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fs := subtest.FileSystemService.DoesLabelRequireUnmount()
			utils.CheckOutput("fss.DoesLabelRequireUnmount()", t, subtest.ExpectedOutput, fs)
		})
	}
}

func TestGetFileSystemSize(t *testing.T) {
	subtests := []struct {
		Name              string
		Device            string
		FileSystemService func(rf utils.RunnerFactory) FileSystemService
		RunnerBinary      utils.Binary
		RunnerArgs        []string
		RunnerOutput      string
		RunnerOutputFile  string
		RunnerError       error
		ExpectedOutput    uint64
		ExpectedError     error
	}{
		{
			Name:              "success<ext4>",
			Device:            "/dev/vdb",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.Tune2fs,
			RunnerArgs:        []string{"-l", "/dev/vdb"},
			RunnerOutputFile:  "testdata/tune2fs.txt",
			RunnerError:       nil,
			ExpectedOutput:    2621440 * 4096, // Block Count * Block Size
			ExpectedError:     nil,
		},
		{
			Name:              "failure<ext4>",
			Device:            "/dev/vdb",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewExt4Service(rf) },
			RunnerBinary:      utils.Tune2fs,
			RunnerArgs:        []string{"-l", "/dev/vdb"},
			RunnerOutput:      "",
			RunnerError:       fmt.Errorf("🔴 tune2fs: No such file or directory while trying to open /dev/vdb"),
			ExpectedOutput:    0,
			ExpectedError:     fmt.Errorf("🔴 tune2fs: No such file or directory while trying to open /dev/vdb"),
		},
		{
			Name:              "success<xfs>",
			Device:            "/dev/vdc",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsInfo,
			RunnerArgs:        []string{"/dev/vdc"},
			RunnerOutputFile:  "testdata/xfs_info.txt",
			RunnerError:       nil,
			ExpectedOutput:    2621440 * 4096, // data(Block Count) * data(Block Size)
			ExpectedError:     nil,
		},
		{
			Name:              "failure<xfs>",
			Device:            "/dev/vdc",
			FileSystemService: func(rf utils.RunnerFactory) FileSystemService { return NewXfsService(rf) },
			RunnerBinary:      utils.XfsInfo,
			RunnerArgs:        []string{"/dev/vdc"},
			RunnerOutput:      "",
			RunnerError:       fmt.Errorf("🔴 /dev/vdc: No such file or directory"),
			ExpectedOutput:    0,
			ExpectedError:     fmt.Errorf("🔴 /dev/vdc: No such file or directory"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			var runnerOutput string
			if len(subtest.RunnerOutputFile) > 0 {
				data, err := os.ReadFile(subtest.RunnerOutputFile)
				utils.ExpectErr("os.ReadFile()", t, false, err)
				runnerOutput = string(data)
			} else {
				runnerOutput = subtest.RunnerOutput
			}

			mrf := utils.NewMockRunnerFactory(subtest.RunnerBinary, subtest.RunnerArgs, runnerOutput, subtest.RunnerError)
			fss := subtest.FileSystemService(mrf)
			size, err := fss.GetSize(subtest.Device)
			utils.CheckError("fss.GetSize()", t, subtest.ExpectedError, err)
			utils.CheckOutput("fss.GetSize()", t, subtest.ExpectedOutput, size)
		})
	}
}
