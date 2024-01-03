package action

import "github.com/reecetech/ebs-bootstrap/internal/model"

type MockAction struct {
	execute func() error
	mode    model.Mode
}

func (ma *MockAction) Execute() error {
	return ma.execute()
}

func (ma *MockAction) Success() string {
	return "Successfully executed action"
}

func (ma *MockAction) Prompt() string {
	return "Would you like to execute action"
}

func (ma *MockAction) Refuse() string {
	return "Refused to execute action"
}

func (ma *MockAction) GetMode() model.Mode {
	return ma.mode
}

func (ma *MockAction) SetMode(mode model.Mode) Action {
	ma.mode = mode
	return ma
}
