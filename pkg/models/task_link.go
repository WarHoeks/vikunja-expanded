// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"io"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskLink is a reference link (repo, prod environment, docs, ...) attached to a task.
type TaskLink struct {
	// The unique, numeric id of this link.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"link" readOnly:"true" doc:"The unique, numeric id of this link."`
	// The id of the task this link belongs to. Taken from the URL, not the body.
	TaskID int64 `xorm:"bigint not null index" json:"task_id" param:"task" readOnly:"true" doc:"The id of the task this link belongs to. Taken from the URL, not the body."`
	// The target URL.
	URL string `xorm:"varchar(2000) not null" json:"url" valid:"required,url,runelength(1|2000)" minLength:"1" maxLength:"2000" doc:"The target URL, e.g. a repository, environment, or docs page."`
	// The display text for this link.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The display text for this link."`
	// A simple-icons (https://simple-icons.org) slug, e.g. "github". Empty when custom_icon_id is set instead.
	Icon string `xorm:"varchar(100) null" json:"icon" maxLength:"100" doc:"A simple-icons (https://simple-icons.org) slug, e.g. \"github\". Empty when custom_icon_id is set instead."`
	// The id of an uploaded custom icon file. 0 when icon (a simple-icons slug) is set instead.
	CustomIconID int64 `xorm:"bigint null" json:"custom_icon_id" readOnly:"true" doc:"The id of an uploaded custom icon file, set via the icon upload endpoint. 0 when icon (a simple-icons slug) is set instead."`

	CreatedByID int64 `xorm:"bigint not null" json:"-"`
	// The user who added this link.
	CreatedBy *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who added this link."`

	// A timestamp when this link was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this link was created. You cannot change this value."`
	// A timestamp when this link was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this link was last updated. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*TaskLink) TableName() string {
	return "task_links"
}

func (tl *TaskLink) validate() error {
	if tl.Icon != "" && tl.CustomIconID != 0 {
		return InvalidFieldErrorWithMessage([]string{"icon", "custom_icon_id"}, "a link can use a simple-icons slug or a custom icon, not both")
	}
	if !strings.HasPrefix(tl.URL, "http://") && !strings.HasPrefix(tl.URL, "https://") {
		return InvalidFieldErrorWithMessage([]string{"url"}, "url must start with http:// or https://")
	}
	return nil
}

// Create creates a new task link
func (tl *TaskLink) Create(s *xorm.Session, a web.Auth) (err error) {
	if err = tl.validate(); err != nil {
		return err
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	tl.ID = 0
	tl.CreatedBy = u
	tl.CreatedByID = u.ID

	_, err = s.Insert(tl)
	return err
}

// ReadAll returns all links for a task
func (tl *TaskLink) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	task := Task{ID: tl.TaskID}
	canRead, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	listCond := builder.Eq{"task_id": tl.TaskID}

	links := []*TaskLink{}
	err = s.Where(listCond).
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&links)
	if err != nil {
		return nil, 0, 0, err
	}

	total, err := s.Where(listCond).Count(&TaskLink{})
	if err != nil {
		return nil, 0, 0, err
	}

	userIDs := make([]int64, 0, len(links))
	for _, l := range links {
		userIDs = append(userIDs, l.CreatedByID)
	}
	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	for _, l := range links {
		if createdBy, has := users[l.CreatedByID]; has {
			l.CreatedBy = createdBy
		}
	}

	return links, len(links), total, nil
}

// Update updates a task link
func (tl *TaskLink) Update(s *xorm.Session, a web.Auth) (err error) {
	if err = tl.validate(); err != nil {
		return err
	}

	_, err = s.
		Where("id = ? AND task_id = ?", tl.ID, tl.TaskID).
		Cols("url", "title", "icon", "custom_icon_id").
		Update(tl)
	if err != nil {
		return err
	}

	u, err := user.GetUserByID(s, tl.CreatedByID)
	if err != nil {
		return err
	}
	tl.CreatedBy = u
	return nil
}

// Delete removes a task link
func (tl *TaskLink) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.
		Where("id = ? AND task_id = ?", tl.ID, tl.TaskID).
		Delete(&TaskLink{})
	return err
}

// loadExisting fetches the current row for tl.ID/tl.TaskID into tl, without the
// auto-condition xorm would otherwise add from tl's already-set fields.
func (tl *TaskLink) loadExisting(s *xorm.Session) error {
	exists, err := s.Where("id = ? AND task_id = ?", tl.ID, tl.TaskID).NoAutoCondition().Get(tl)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskLinkDoesNotExist{ID: tl.ID, TaskID: tl.TaskID}
	}
	return nil
}

// SetCustomIcon uploads f as this link's custom icon, replacing any simple-icons
// slug or previously uploaded custom icon. The caller owns the session and commit.
func (tl *TaskLink) SetCustomIcon(s *xorm.Session, f io.ReadSeeker, realname string, realsize uint64, a web.Auth) error {
	if err := tl.loadExisting(s); err != nil {
		return err
	}
	oldCustomIconID := tl.CustomIconID

	file, err := files.CreateWithSession(s, f, realname, realsize, a)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return ErrTaskLinkIconIsTooLarge{Size: realsize}
		}
		return err
	}

	tl.CustomIconID = file.ID
	tl.Icon = ""
	if _, err := s.ID(tl.ID).Cols("custom_icon_id", "icon").Update(tl); err != nil {
		_ = file.Delete(s)
		return err
	}

	if oldCustomIconID > 0 {
		if err := (&files.File{ID: oldCustomIconID}).Delete(s); err != nil && !files.IsErrFileDoesNotExist(err) {
			return err
		}
	}

	return nil
}

// GetTaskLinkIconFile loads the file for a link's custom icon, checking read
// access to the parent task first. Owns its own session, committed before the
// storage read so no pool connection is held while streaming.
func GetTaskLinkIconFile(a web.Auth, taskID, linkID int64) (f *files.File, err error) {
	s := db.NewSession()
	defer s.Close()

	tl := &TaskLink{ID: linkID, TaskID: taskID}
	can, _, err := tl.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}

	if err := tl.loadExisting(s); err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if tl.CustomIconID == 0 {
		_ = s.Rollback()
		return nil, ErrTaskLinkHasNoCustomIcon{ID: tl.ID}
	}

	f = &files.File{ID: tl.CustomIconID}
	if err := f.LoadFileMetaByID(s); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := f.LoadFileByID(); err != nil {
		return nil, err
	}

	return f, nil
}
