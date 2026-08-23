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

// ProjectLink is a reference link (repo, prod environment, docs, ...) attached to a project.
type ProjectLink struct {
	// The unique, numeric id of this link.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"link" readOnly:"true" doc:"The unique, numeric id of this link."`
	// The id of the project this link belongs to. Taken from the URL, not the body.
	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project" readOnly:"true" doc:"The id of the project this link belongs to. Taken from the URL, not the body."`
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
func (*ProjectLink) TableName() string {
	return "project_links"
}

func (pl *ProjectLink) validate() error {
	if pl.Icon != "" && pl.CustomIconID != 0 {
		return InvalidFieldErrorWithMessage([]string{"icon", "custom_icon_id"}, "a link can use a simple-icons slug or a custom icon, not both")
	}
	if !strings.HasPrefix(pl.URL, "http://") && !strings.HasPrefix(pl.URL, "https://") {
		return InvalidFieldErrorWithMessage([]string{"url"}, "url must start with http:// or https://")
	}
	return nil
}

// Create creates a new project link
func (pl *ProjectLink) Create(s *xorm.Session, a web.Auth) (err error) {
	if err = pl.validate(); err != nil {
		return err
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	pl.ID = 0
	pl.CreatedBy = u
	pl.CreatedByID = u.ID

	_, err = s.Insert(pl)
	return err
}

// ReadAll returns all links for a project
func (pl *ProjectLink) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	p := &Project{ID: pl.ProjectID}
	can, _, err := p.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	listCond := builder.Eq{"project_id": pl.ProjectID}

	links := []*ProjectLink{}
	err = s.Where(listCond).
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&links)
	if err != nil {
		return nil, 0, 0, err
	}

	total, err := s.Where(listCond).Count(&ProjectLink{})
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

// Update updates a project link
func (pl *ProjectLink) Update(s *xorm.Session, a web.Auth) (err error) {
	if err = pl.validate(); err != nil {
		return err
	}

	_, err = s.
		Where("id = ? AND project_id = ?", pl.ID, pl.ProjectID).
		Cols("url", "title", "icon", "custom_icon_id").
		Update(pl)
	if err != nil {
		return err
	}

	u, err := user.GetUserByID(s, pl.CreatedByID)
	if err != nil {
		return err
	}
	pl.CreatedBy = u
	return nil
}

// Delete removes a project link
func (pl *ProjectLink) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.
		Where("id = ? AND project_id = ?", pl.ID, pl.ProjectID).
		Delete(&ProjectLink{})
	return err
}

// loadExisting fetches the current row for pl.ID/pl.ProjectID into pl, without the
// auto-condition xorm would otherwise add from pl's already-set fields.
func (pl *ProjectLink) loadExisting(s *xorm.Session) error {
	exists, err := s.Where("id = ? AND project_id = ?", pl.ID, pl.ProjectID).NoAutoCondition().Get(pl)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProjectLinkDoesNotExist{ID: pl.ID, ProjectID: pl.ProjectID}
	}
	return nil
}

// SetCustomIcon uploads f as this link's custom icon, replacing any simple-icons
// slug or previously uploaded custom icon. The caller owns the session and commit.
func (pl *ProjectLink) SetCustomIcon(s *xorm.Session, f io.ReadSeeker, realname string, realsize uint64, a web.Auth) error {
	if err := pl.loadExisting(s); err != nil {
		return err
	}
	oldCustomIconID := pl.CustomIconID

	file, err := files.CreateWithSession(s, f, realname, realsize, a)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return ErrProjectLinkIconIsTooLarge{Size: realsize}
		}
		return err
	}

	pl.CustomIconID = file.ID
	pl.Icon = ""
	if _, err := s.ID(pl.ID).Cols("custom_icon_id", "icon").Update(pl); err != nil {
		_ = file.Delete(s)
		return err
	}

	if err := deleteCustomIconFileIfUnused(s, oldCustomIconID); err != nil {
		return err
	}

	return nil
}

// SetCustomIconFromLibrary attaches an existing custom icon library entry to this
// link by copying its file reference — no new upload, and the file may end up
// shared by several links. The caller owns the session and commit.
func (pl *ProjectLink) SetCustomIconFromLibrary(s *xorm.Session, customIconID int64, a web.Auth) error {
	if err := pl.loadExisting(s); err != nil {
		return err
	}
	oldCustomIconID := pl.CustomIconID

	fileID, err := GetCustomIconFileID(s, a, customIconID)
	if err != nil {
		return err
	}

	pl.CustomIconID = fileID
	pl.Icon = ""
	if _, err := s.ID(pl.ID).Cols("custom_icon_id", "icon").Update(pl); err != nil {
		return err
	}

	return deleteCustomIconFileIfUnused(s, oldCustomIconID)
}

// GetProjectLinkIconFile loads the file for a link's custom icon, checking read
// access to the parent project first. Owns its own session, committed before the
// storage read so no pool connection is held while streaming.
func GetProjectLinkIconFile(a web.Auth, projectID, linkID int64) (f *files.File, err error) {
	s := db.NewSession()
	defer s.Close()

	pl := &ProjectLink{ID: linkID, ProjectID: projectID}
	can, _, err := pl.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}

	if err := pl.loadExisting(s); err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if pl.CustomIconID == 0 {
		_ = s.Rollback()
		return nil, ErrProjectLinkHasNoCustomIcon{ID: pl.ID}
	}

	f = &files.File{ID: pl.CustomIconID}
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
