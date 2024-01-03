package utils

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ryanuber/go-glob"
)

type NotImeplementedError struct {
	Id string
}

func NewNotImeplementedError(id string) *NotImeplementedError {
	return &NotImeplementedError{
		Id: id,
	}
}

func (e *NotImeplementedError) Error() string {
	return fmt.Sprintf("🔴 %s is not implemented", e.Id)
}

func CheckOutput(id string, t *testing.T, expected interface{}, actual interface{}, opts ...cmp.Option) {
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("%s [output] mismatch: Expected=%+v Actual=%+v", id, expected, actual)
	}
}

func ExpectErr(id string, t *testing.T, shouldExpectErr bool, err error) {
	if shouldExpectErr && err == nil {
		t.Fatalf("%s [error] mismatch: No error detected, despite it being expected", id)
	}
	if !shouldExpectErr && err != nil {
		t.Fatalf("%s [error] mismatch: Expected=%v, Actual=%s", id, nil, err)
	}
}

func CheckError(id string, t *testing.T, expected error, actual error) {
	if expected != nil {
		if actual == nil {
			t.Fatalf("%s [error] undetected: Expected=%v", id, expected)
			return
		}
		if expected.Error() != actual.Error() {
			t.Fatalf("%s [error] mismatch: Expected=%v Actual=%v", id, expected, actual)
		}
	}
	if actual != nil {
		if expected == nil {
			t.Fatalf("%s [error] undetected: Actual=%v", id, actual)
			return
		}
	}
}

func CheckErrorGlob(id string, t *testing.T, pattern error, actual error) {
	if pattern != nil {
		if actual == nil {
			t.Fatalf("%s [error] undetected: Pattern=%v", id, pattern)
			return
		}
		// Perform a glob match of the error message. Glob matching is useful
		// for error messages that contain dynamic attributes
		if !glob.Glob(pattern.Error(), actual.Error()) {
			t.Fatalf("%s [error] mismatch: Pattern=%v Actual=%v", id, pattern, actual)
		}
	}
	if actual != nil {
		if pattern == nil {
			t.Fatalf("%s [error] undetected: Actual=%v", id, actual)
			return
		}
	}
}

type MockIncrementErrorType string

const (
	ErrorUntilTrigger   MockIncrementErrorType = "ErrorUntilTrigger"
	SuccessUntilTrigger MockIncrementErrorType = "SuccessUntilTrigger"
)

type MockIncrementError struct {
	id            string
	mockErrorType MockIncrementErrorType
	trigger       uint32
	increment     uint32
}

func NewMockIncrementError(id string, errorType MockIncrementErrorType, trigger uint32) *MockIncrementError {
	return &MockIncrementError{
		id:            id,
		mockErrorType: errorType,
		trigger:       trigger,
	}
}

func (mie *MockIncrementError) Error() string {
	return fmt.Sprintf("🔴 %s: Type=%s, Increment=%d, Trigger=%d", mie.id, mie.mockErrorType, mie.increment, mie.trigger)
}

func (mie *MockIncrementError) Trigger() error {
	mie.increment++
	switch mie.mockErrorType {
	case ErrorUntilTrigger:
		if mie.increment < mie.trigger {
			return mie
		}
	case SuccessUntilTrigger:
		if mie.increment >= mie.trigger {
			return mie
		}
	default:
		return fmt.Errorf("🔴 An unsupported MockLayerErrorType was encountered")
	}
	return nil
}
