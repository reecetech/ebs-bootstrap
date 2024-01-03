package service

import (
	"os"
	"path"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetFile(t *testing.T) {
	file, err := regularFile()
	defer os.Remove(file)
	utils.CheckError("temporaryFile()", t, nil, err)

	dir, err := directory()
	defer os.RemoveAll(dir)
	utils.CheckError("temporaryDirectory()", t, nil, err)

	// Source/Destination Regular File Symlink (srfs/drfs)
	srfs, drfs, err := symlink(model.RegularFile)
	utils.CheckError("symlink(model.RegularFile)", t, nil, err)
	defer os.RemoveAll(path.Dir(drfs))

	// Source/Destination Directory Symlink (sds/dds)
	sds, dds, err := symlink(model.Directory)
	utils.CheckError("symlink(model.Directory)", t, nil, err)
	defer os.RemoveAll(path.Dir(dds))

	// Source/Destination Special Symlink (sss/dss)
	sss, dss, err := symlink(model.Special)
	utils.CheckError("symlink(model.Special)", t, nil, err)
	defer os.RemoveAll(path.Dir(dss))

	ufs := NewUnixFileService()

	subtests := []struct {
		Name            string
		Path            string
		ExpectedPath    string
		ExpectedType    model.FileType
		ShouldExpectErr bool
	}{
		{
			Name:            "Regular File",
			Path:            file,
			ExpectedPath:    file,
			ExpectedType:    model.RegularFile,
			ShouldExpectErr: false,
		},
		{
			Name:            "Directory",
			Path:            dir,
			ExpectedPath:    dir,
			ExpectedType:    model.Directory,
			ShouldExpectErr: false,
		},
		{
			Name:            "Symlink<Regular File>",
			Path:            drfs,
			ExpectedPath:    srfs,
			ExpectedType:    model.RegularFile,
			ShouldExpectErr: false,
		},
		{
			Name:            "Symlink<Directory>",
			Path:            dds,
			ExpectedPath:    sds,
			ExpectedType:    model.Directory,
			ShouldExpectErr: false,
		},
		{
			Name:            "Symlink<Special>",
			Path:            dss,
			ExpectedPath:    sss,
			ExpectedType:    model.Special,
			ShouldExpectErr: false,
		},
		{
			Name:            "File That Does Not Exist",
			Path:            "/dev/does-not-exit",
			ExpectedPath:    "",
			ExpectedType:    0,
			ShouldExpectErr: true,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			output, err := ufs.GetFile(subtest.Path)
			utils.ExpectErr("ufs.GetFile()", t, subtest.ShouldExpectErr, err)
			utils.CheckOutput("ufs.GetFile()", t, subtest.ExpectedPath, utils.Safe(output).Path)
			utils.CheckOutput("ufs.GetFile()", t, subtest.ExpectedType, utils.Safe(output).Type)
		})
	}
}

func TestDirectoryModifications(t *testing.T) {
	ufs := NewUnixFileService()

	// Create Temporary Directory
	dir, err := directory()
	utils.ExpectErr("temporaryDirectory()", t, false, err)
	defer os.RemoveAll(dir)

	// Create Nested Directory Inside Temporary Directory
	nested := path.Join(dir, "nested")
	err = ufs.CreateDirectory(nested)
	utils.ExpectErr("ufs.CreateDirectory()", t, false, err)

	// Get File Information of Nested Directory
	file, err := ufs.GetFile(nested)
	utils.ExpectErr("ufs.GetFile()", t, false, err)

	// Change Permissions of Nested Directory
	err = ufs.ChangePermissions(nested, file.Permissions)
	utils.ExpectErr("ufs.ChangePermissions()", t, false, err)

	// Change Owner of Nested Directory to Match Temporary Directory
	err = ufs.ChangeOwner(nested, file.UserId, file.GroupId)
	utils.ExpectErr("ufs.ChangeOwner()", t, false, err)
}

// Create a temporary file
func regularFile() (string, error) {
	file, err := os.CreateTemp("", "temp_file")
	if err != nil {
		return "", err
	}
	defer file.Close()
	return file.Name(), nil
}

// Create a temporary directory
func directory() (string, error) {
	dir, err := os.MkdirTemp("", "temp_dir")
	if err != nil {
		return "", err
	}
	return dir, nil
}

// Create a temporary symbolic link
func symlink(ft model.FileType) (string, string, error) {
	dir, err := os.MkdirTemp("", "temp_symlink")
	if err != nil {
		return "", "", err
	}
	var src string
	switch ft {
	case model.RegularFile:
		f, err := os.CreateTemp(dir, "source")
		if err != nil {
			return "", "", nil
		}
		defer f.Close()
		src = f.Name()
	case model.Directory:
		src, err = os.MkdirTemp(dir, "source")
		if err != nil {
			return "", "", nil
		}
	default:
		src = "/dev/null"
	}
	dst := path.Join(dir, "destination")
	err = os.Symlink(src, dst)
	if err != nil {
		return "", "", nil
	}
	return src, dst, nil
}
