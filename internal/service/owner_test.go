package service

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetUser(t *testing.T) {
	uos := NewUnixOwnerService()

	cu, err := uos.GetCurrentUser()
	utils.ExpectErr("uos.GetCurrentUser()", t, false, err)

	subtests := []struct {
		Name           string
		User           string
		ExpectedOutput *model.User
		ExpectedError  error
	}{
		{
			Name: "Current User + Username",
			User: cu.Name,
			ExpectedOutput: &model.User{
				Name: cu.Name,
				Id:   cu.Id,
			},
			ExpectedError: nil,
		},
		{
			Name: "Current User + User Id",
			User: fmt.Sprintf("%d", cu.Id),
			ExpectedOutput: &model.User{
				Name: cu.Name,
				Id:   cu.Id,
			},
			ExpectedError: nil,
		},
		{
			Name:           "Username That Does Not Exist",
			User:           "inva1id",
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 User (name=inva1id) does not exist"),
		},
		{
			Name:           "User Id That Does Not Exist",
			User:           "-1",
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 User (id=-1) does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			user, err := uos.GetUser(subtest.User)
			utils.CheckError("uos.GetUser()", t, subtest.ExpectedError, err)
			utils.CheckOutput("uos.GetUser()", t, subtest.ExpectedOutput, user)
		})
	}
}

func TestGetGroup(t *testing.T) {
	uos := NewUnixOwnerService()

	cg, err := uos.GetCurrentGroup()
	utils.ExpectErr("uos.GetCurrentUser()", t, false, err)

	subtests := []struct {
		Name           string
		Group          string
		ExpectedOutput *model.Group
		ExpectedError  error
	}{
		{
			Name:  "Current User + Group Name",
			Group: cg.Name,
			ExpectedOutput: &model.Group{
				Name: cg.Name,
				Id:   cg.Id,
			},
			ExpectedError: nil,
		},
		{
			Name:  "Current User + Group Id",
			Group: fmt.Sprintf("%d", cg.Id),
			ExpectedOutput: &model.Group{
				Name: cg.Name,
				Id:   cg.Id,
			},
			ExpectedError: nil,
		},
		{
			Name:           "Group That Does Not Exist",
			Group:          "inva1id",
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 Group (name=inva1id) does not exist"),
		},
		{
			Name:           "Group Id That Does Not Exist",
			Group:          "-1",
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 Group (id=-1) does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			user, err := uos.GetGroup(subtest.Group)
			utils.CheckError("uos.GetGroup()", t, subtest.ExpectedError, err)
			utils.CheckOutput("uos.GetGroup()", t, subtest.ExpectedOutput, user)
		})
	}
}
