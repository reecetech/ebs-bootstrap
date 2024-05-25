package layer

import (
	"fmt"
	"testing"
	"time"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

const (
	MaxUint32 = ^uint32(0)
)

type MockLayer struct {
	from          *utils.MockIncrementError
	modify        *utils.MockIncrementError
	validate      *utils.MockIncrementError
	shouldProcess bool
}

func (ml *MockLayer) From(c *config.Config) error {
	return ml.from.Trigger()
}

func (ml *MockLayer) Modify(c *config.Config) ([]action.Action, error) {
	err := ml.modify.Trigger()
	if err != nil {
		return nil, err
	}
	return []action.Action{}, nil
}

func (ml *MockLayer) Validate(c *config.Config) error {
	return ml.validate.Trigger()
}

func (ml *MockLayer) Warning() string {
	return DisabledWarning
}

func (ml *MockLayer) ShouldProcess(c *config.Config) bool {
	return ml.shouldProcess
}

func TestExponentialBackoffLayerExecutor(t *testing.T) {
	mae := action.NewDefaultActionExecutor()
	// Lets generate ExponentialBackoffParameters with a custom
	// InitialInterval of 10 ms. We do not want to slow down the test suite
	// with an excessively long initial interval
	debp := DefaultExponentialBackoffParameters()
	ebp := &ExponentialBackoffParameters{
		InitialInterval: 10 * time.Millisecond,
		Multiplier:      debp.Multiplier,
		MaxRetries:      debp.MaxRetries,
	}

	subtests := []struct {
		Name          string
		From          *utils.MockIncrementError
		Modify        *utils.MockIncrementError
		Validate      *utils.MockIncrementError
		ExpectedError error
	}{
		{
			Name:          "Success",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, MaxUint32),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, MaxUint32),
			Validate:      utils.NewMockIncrementError("Validate()", utils.SuccessUntilTrigger, MaxUint32),
			ExpectedError: nil,
		},
		{
			Name:          "From() - Failure on First Call",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, 1),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, MaxUint32),
			Validate:      utils.NewMockIncrementError("Validate()", utils.SuccessUntilTrigger, MaxUint32),
			ExpectedError: fmt.Errorf("🔴 From(): Type=SuccessUntilTrigger, Increment=1, Trigger=1"),
		},
		{
			Name:          "From() - Trigger Permanent Backoff Failure",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, 2),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, MaxUint32),
			Validate:      utils.NewMockIncrementError("Validate()", utils.SuccessUntilTrigger, MaxUint32),
			ExpectedError: fmt.Errorf("🔴 From(): Type=SuccessUntilTrigger, Increment=2, Trigger=2"),
		},
		{
			Name:          "Modify() - Failure on First Call",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, MaxUint32),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, 1),
			Validate:      utils.NewMockIncrementError("Validate()", utils.SuccessUntilTrigger, MaxUint32),
			ExpectedError: fmt.Errorf("🔴 Modify(): Type=SuccessUntilTrigger, Increment=1, Trigger=1"),
		},
		// The number of times Validate() would be the initial call (1) plus the number of allowed retries (ebp.MaxRetries)
		{
			Name:          "Validate() - Success Just Before Maximum Retries Reached",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, MaxUint32),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, MaxUint32),
			Validate:      utils.NewMockIncrementError("Validate()", utils.ErrorUntilTrigger, 1+ebp.MaxRetries),
			ExpectedError: nil,
		},
		{
			Name:          "Validate() - Trigger Maximum Retries",
			From:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, MaxUint32),
			Modify:        utils.NewMockIncrementError("Modify()", utils.SuccessUntilTrigger, MaxUint32),
			Validate:      utils.NewMockIncrementError("Validate()", utils.ErrorUntilTrigger, 2+ebp.MaxRetries),
			ExpectedError: fmt.Errorf("🔴 Validate(): Type=ErrorUntilTrigger, Increment=%d, Trigger=%d", 1+ebp.MaxRetries, 2+ebp.MaxRetries),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ml := &MockLayer{
				from:          subtest.From,
				modify:        subtest.Modify,
				validate:      subtest.Validate,
				shouldProcess: true,
			}
			eb := NewExponentialBackoffLayerExecutor(nil, mae, ebp)
			err := eb.Execute([]Layer{ml})
			utils.CheckError("eb.Execute()", t, subtest.ExpectedError, err)
		})
	}
}

func TestExponentialBackoffLayerExecutorShouldProcess(t *testing.T) {
	// This MockLayer will fail if any of the From(), Modify() or Validate() methods
	// are invoked. `shouldProcess` has been set to false so these methods would be skipped
	// and there should be no error generated
	ml := &MockLayer{
		from:          utils.NewMockIncrementError("From()", utils.SuccessUntilTrigger, 1),
		modify:        utils.NewMockIncrementError("Modidy()", utils.SuccessUntilTrigger, 1),
		validate:      utils.NewMockIncrementError("Validate()", utils.SuccessUntilTrigger, 1),
		shouldProcess: false,
	}
	debp := DefaultExponentialBackoffParameters()
	eb := NewExponentialBackoffLayerExecutor(nil, nil, debp)
	err := eb.Execute([]Layer{ml})
	utils.CheckError("eb.Execute()", t, nil, err)
}
