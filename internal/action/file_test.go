package action

import (
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestCreateDirectoryActionExecute(t *testing.T) {
	mfs := service.NewMockFileService()
	mfs.StubCreateDirectory = func(p string) error { return nil }
	cda := NewCreateDirectoryAction("/mnt/foo", mfs)
	utils.ExpectErr("cda.Execute()", t, false, cda.Execute())
}

func TestCreateDirectoryActionMode(t *testing.T) {
	cda := NewCreateDirectoryAction("/mnt/foo", nil)
	cda.SetMode(model.Healthcheck)
	utils.CheckOutput("cda.GetMode()", t, model.Healthcheck, cda.GetMode())
}

func TestCreateDirectoryActionMessages(t *testing.T) {
	cda := NewCreateDirectoryAction("/mnt/foo", nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        cda.Prompt(),
			ExpectedOutput: "Would you like to recursively create directory /mnt/foo",
		},
		{
			Name:           "Refuse",
			Message:        cda.Refuse(),
			ExpectedOutput: "Refused to create directory /mnt/foo",
		},
		{
			Name:           "Success",
			Message:        cda.Success(),
			ExpectedOutput: "Successfully created directory /mnt/foo",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}

func TestChangeOwnerActionExecute(t *testing.T) {
	mfs := service.NewMockFileService()
	mfs.StubChangeOwner = func(p string, uid model.UserId, gid model.GroupId) error { return nil }
	coa := NewChangeOwnerAction("/mnt/foo", 0, 0, mfs)
	utils.ExpectErr("cda.Execute()", t, false, coa.Execute())
}

func TestChangeOwnerActionMode(t *testing.T) {
	coa := NewChangeOwnerAction("/mnt/foo", 0, 0, nil)
	coa.SetMode(model.Healthcheck)
	utils.CheckOutput("cda.GetMode()", t, model.Healthcheck, coa.GetMode())
}

func TestChangeOwnerActionMessages(t *testing.T) {
	coa := NewChangeOwnerAction("/mnt/foo", 1000, 2000, nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        coa.Prompt(),
			ExpectedOutput: "Would you like to change ownership (1000:2000) of /mnt/foo",
		},
		{
			Name:           "Refuse",
			Message:        coa.Refuse(),
			ExpectedOutput: "Refused to to change ownership (1000:2000) of /mnt/foo",
		},
		{
			Name:           "Success",
			Message:        coa.Success(),
			ExpectedOutput: "Successfully changed ownership (1000:2000) of /mnt/foo",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}

func TestChangePermissionsActionExecute(t *testing.T) {
	mfs := service.NewMockFileService()
	mfs.StubChangePermissions = func(p string, perms model.FilePermissions) error { return nil }
	cpa := NewChangePermissionsAction("/mnt/foo", 0644, mfs)
	utils.ExpectErr("cpa.Execute()", t, false, cpa.Execute())
}

func TestChangePermissionsActionMode(t *testing.T) {
	cpa := NewChangePermissionsAction("/mnt/foo", 0644, nil)
	cpa.SetMode(model.Healthcheck)
	utils.CheckOutput("cpa.GetMode()", t, model.Healthcheck, cpa.GetMode())
}

func TestChangePermissionsActionMessages(t *testing.T) {
	cpa := NewChangePermissionsAction("/mnt/foo", 0755, nil)
	subtests := []struct {
		Name           string
		Message        string
		ExpectedOutput string
	}{
		{
			Name:           "Prompt",
			Message:        cpa.Prompt(),
			ExpectedOutput: "Would you like to change permissions of /mnt/foo to 0755",
		},
		{
			Name:           "Refuse",
			Message:        cpa.Refuse(),
			ExpectedOutput: "Refused to to change permissions of /mnt/foo to 0755",
		},
		{
			Name:           "Success",
			Message:        cpa.Success(),
			ExpectedOutput: "Successfully change permissions of /mnt/foo to 0755",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput(subtest.Name, t, subtest.ExpectedOutput, subtest.Message)
		})
	}
}
