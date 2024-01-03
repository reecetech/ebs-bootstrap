package action

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestDefaultActionExecutor(t *testing.T) {
	var input = func(i string) func(buffer *string) error {
		return func(buffer *string) error {
			*buffer = i
			return nil
		}
	}
	var fail = func(err string) func(buffer *string) error {
		return func(buffer *string) error {
			return fmt.Errorf(err)
		}
	}

	subtests := []struct {
		Name          string
		Read          func(buffer *string) error
		Error         error
		Mode          model.Mode
		ExpectedError error
	}{
		{
			Name:          "Mode=Empty + Read=Disabled + Action=Success",
			Read:          fail("🔴 Standard Input Disabled"),
			Error:         nil,
			Mode:          model.Empty,
			ExpectedError: fmt.Errorf("🔴 Unsupported mode was encountered. Refused to execute action"),
		},
		{
			Name:          "Mode=Healthcheck + Read=Avoid + Action=Success",
			Read:          fail("🔴 Standard Input Disabled"),
			Error:         nil,
			Mode:          model.Healthcheck,
			ExpectedError: fmt.Errorf("🔴 Healthcheck mode enabled. Refused to execute action"),
		},
		{
			Name:          "Mode=Prompt + Read=Input<y> + Action=Success",
			Read:          input("y"),
			Error:         nil,
			Mode:          model.Prompt,
			ExpectedError: nil,
		},
		{
			Name:          "Mode=Prompt + Read=Input<yes> + Action=Success",
			Read:          input("yes"),
			Error:         nil,
			Mode:          model.Prompt,
			ExpectedError: nil,
		},
		{
			Name:          "Mode=Prompt + Read=Failure + Action=Success",
			Read:          fail("🔴 Failed to Read From Standard Input"),
			Error:         nil,
			Mode:          model.Prompt,
			ExpectedError: fmt.Errorf("🔴 Action rejected. Refused to execute action"),
		},
		{
			Name:          "Mode=Prompt + Read=Input<n> + Action=Success",
			Read:          input("n"),
			Error:         nil,
			Mode:          model.Prompt,
			ExpectedError: fmt.Errorf("🔴 Action rejected. Refused to execute action"),
		},
		{
			Name:          "Mode=Force + Read=Disabled + Action=Success",
			Read:          fail("🔴 Standard Input Disabled"),
			Error:         nil,
			Mode:          model.Force,
			ExpectedError: nil,
		},
		{
			Name:          "Mode=Force + Read=Disabled + Action=Failure",
			Read:          fail("🔴 Standard Input Disabled"),
			Error:         fmt.Errorf("🔴 Error encountered while executing action"),
			Mode:          model.Force,
			ExpectedError: fmt.Errorf("🔴 Error encountered while executing action"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			dae := &DefaultActionExecutor{
				read: subtest.Read,
			}
			a := (&MockAction{
				execute: func() error { return subtest.Error },
			})
			err := dae.Execute([]Action{a.SetMode(subtest.Mode)})
			utils.CheckError("dae.Execute()", t, subtest.ExpectedError, err)
		})
	}
}
