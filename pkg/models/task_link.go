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
	"strings"
	"time"

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

	// tl.CreatedByID is never set from the request body (json:"-", and Update's
	// own Cols() list deliberately excludes it too — the creator doesn't change
	// on edit) — it's zero here. Load the real value from the stored row into a
	// separate struct so we don't clobber tl's new URL/Title with the old ones.
	existing := &TaskLink{ID: tl.ID, TaskID: tl.TaskID}
	if err := existing.loadExisting(s); err != nil {
		return err
	}
	tl.CreatedByID = existing.CreatedByID

	_, err = s.
		Where("id = ? AND task_id = ?", tl.ID, tl.TaskID).
		Cols("url", "title").
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
