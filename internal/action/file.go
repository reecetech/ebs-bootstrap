package action

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type CreateDirectoryAction struct {
	path        string
	mode        model.Mode
	fileService service.FileService
}

func NewCreateDirectoryAction(p string, fs service.FileService) *CreateDirectoryAction {
	return &CreateDirectoryAction{
		path:        p,
		mode:        model.Empty,
		fileService: fs,
	}
}

func (a *CreateDirectoryAction) Execute() error {
	return a.fileService.CreateDirectory(a.path)
}

func (a *CreateDirectoryAction) GetMode() model.Mode {
	return a.mode
}

func (a *CreateDirectoryAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *CreateDirectoryAction) Prompt() string {
	return fmt.Sprintf("Would you like to recursively create directory %s", a.path)
}

func (a *CreateDirectoryAction) Refuse() string {
	return fmt.Sprintf("Refused to create directory %s", a.path)
}

func (a *CreateDirectoryAction) Success() string {
	return fmt.Sprintf("Successfully created directory %s", a.path)
}

type ChangeOwnerAction struct {
	path        string
	uid         model.UserId
	gid         model.GroupId
	mode        model.Mode
	fileService service.FileService
}

func NewChangeOwnerAction(p string, uid model.UserId, gid model.GroupId, fs service.FileService) *ChangeOwnerAction {
	return &ChangeOwnerAction{
		path:        p,
		uid:         uid,
		gid:         gid,
		mode:        model.Empty,
		fileService: fs,
	}
}

func (a *ChangeOwnerAction) Execute() error {
	return a.fileService.ChangeOwner(a.path, a.uid, a.gid)
}

func (a *ChangeOwnerAction) GetMode() model.Mode {
	return a.mode
}

func (a *ChangeOwnerAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *ChangeOwnerAction) Prompt() string {
	return fmt.Sprintf("Would you like to change ownership (%d:%d) of %s", a.uid, a.gid, a.path)
}

func (a *ChangeOwnerAction) Refuse() string {
	return fmt.Sprintf("Refused to to change ownership (%d:%d) of %s", a.uid, a.gid, a.path)
}

func (a *ChangeOwnerAction) Success() string {
	return fmt.Sprintf("Successfully changed ownership (%d:%d) of %s", a.uid, a.gid, a.path)
}

type ChangePermissionsAction struct {
	path        string
	perms       model.FilePermissions
	mode        model.Mode
	fileService service.FileService
}

func NewChangePermissionsAction(p string, perms model.FilePermissions, fs service.FileService) *ChangePermissionsAction {
	return &ChangePermissionsAction{
		path:        p,
		perms:       perms,
		mode:        model.Empty,
		fileService: fs,
	}
}

func (a *ChangePermissionsAction) Execute() error {
	return a.fileService.ChangePermissions(a.path, a.perms)
}

func (a *ChangePermissionsAction) GetMode() model.Mode {
	return a.mode
}

func (a *ChangePermissionsAction) SetMode(mode model.Mode) Action {
	a.mode = mode
	return a
}

func (a *ChangePermissionsAction) Prompt() string {
	return fmt.Sprintf("Would you like to change permissions of %s to %#o", a.path, a.perms)
}

func (a *ChangePermissionsAction) Refuse() string {
	return fmt.Sprintf("Refused to to change permissions of %s to %#o", a.path, a.perms)
}

func (a *ChangePermissionsAction) Success() string {
	return fmt.Sprintf("Successfully change permissions of %s to %#o", a.path, a.perms)
}
