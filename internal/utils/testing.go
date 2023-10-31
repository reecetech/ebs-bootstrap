package utils

import (
	"fmt"
	"strings"
	"os/user"
	"testing"
)

func CheckError(name string, t *testing.T, expected error, actual error) {
	if actual != nil {
		if expected == nil {
			t.Errorf("%s [error] undetected: Actual=%v", name, actual)
			return
		}
		if expected.Error() != actual.Error() {
			t.Errorf("%s [error] mismatch: Expected=%v Actual=%v", name, expected, actual)
		}
	}
}

func GetCurrentUserGroup() (*user.User, *user.Group, error) {
	u, err := user.Current()
	if err != nil {
		return nil, nil, fmt.Errorf("🔴 Failed to get current user")
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		return nil, nil, fmt.Errorf("🔴 Failed to get current group")
	}
	/* user.Current() -> From experience, this function can return a username
	   in a capital case. This is not valid UNIX format for usernames so force
	   to lowercase */
	u.Name = strings.ToLower(u.Name)
	g.Name = strings.ToLower(g.Name)
	return u, g, nil
}
