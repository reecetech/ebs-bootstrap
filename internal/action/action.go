package action

import (
	"fmt"
	"log"
	"strings"

	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type Action interface {
	Execute() error
	Success() string
	Prompt() string
	Refuse() string
	GetMode() model.Mode
	SetMode(mode model.Mode) Action
}

type ActionExecutor interface {
	Execute(actions []Action) error
}

type DefaultActionExecutor struct {
	read func(buffer *string) error
}

func NewDefaultActionExecutor() *DefaultActionExecutor {
	return &DefaultActionExecutor{
		read: func(buffer *string) error {
			_, err := fmt.Scanln(buffer)
			return err
		},
	}
}

func (dae *DefaultActionExecutor) Execute(actions []Action) error {
	for _, a := range actions {
		err := dae.execute(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dae *DefaultActionExecutor) execute(action Action) error {
	switch action.GetMode() {
	case model.Force:
		break
	case model.Prompt:
		if !dae.shouldProceed(action) {
			return fmt.Errorf("🔴 Action rejected. %s", action.Refuse())
		}
	case model.Healthcheck:
		return fmt.Errorf("🔴 Healthcheck mode enabled. %s", action.Refuse())
	default:
		return fmt.Errorf("🔴 Unsupported mode was encountered. %s", action.Refuse())
	}

	if err := action.Execute(); err != nil {
		return err
	}
	log.Printf("⭐ %s", action.Success())
	return nil
}

func (dae *DefaultActionExecutor) shouldProceed(action Action) bool {
	prompt := action.Prompt()

	fmt.Printf("🟣 %s? (y/n): ", prompt)
	var response string
	err := dae.read(&response)
	if err != nil {
		return false
	}

	response = strings.ToLower(response)
	if response == "y" || response == "yes" {
		return true
	}
	return false
}
