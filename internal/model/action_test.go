package model

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestParseMode(t *testing.T) {
	subtests := []struct {
		Mode           string
		ExpectedOutput Mode
		ExpectedError  error
	}{
		{
			Mode:           "",
			ExpectedOutput: Empty,
			ExpectedError:  nil,
		},
		{
			Mode:           "healthcheck",
			ExpectedOutput: Healthcheck,
			ExpectedError:  nil,
		},
		{
			Mode:           "prompt",
			ExpectedOutput: Prompt,
			ExpectedError:  nil,
		},
		{
			Mode:           "force",
			ExpectedOutput: Force,
			ExpectedError:  nil,
		},
		{
			Mode:           "invalid",
			ExpectedOutput: Mode("invalid"),
			ExpectedError:  fmt.Errorf("🔴 Mode 'invalid' is not supported"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Mode, func(t *testing.T) {
			m, err := ParseMode(subtest.Mode)
			utils.CheckError("ParseMode()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ParseMode()", t, subtest.ExpectedOutput, m)
		})
	}
}
