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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CustomIcon is a reusable, shared custom icon in the instance-wide icon library that
// project/task links can attach in place of a simple-icons slug.
type CustomIcon struct {
	// The unique, numeric id of this custom icon.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"customicon" readOnly:"true" doc:"The unique, numeric id of this custom icon."`
	// The searchable name of this custom icon.
	Name string `xorm:"varchar(250) not null" json:"name" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The searchable name of this custom icon."`
	// The id of the underlying uploaded file. Not exposed directly; use the icon endpoint to fetch the image.
	FileID int64 `xorm:"bigint not null" json:"-"`

	CreatedByID int64 `xorm:"bigint not null" json:"-"`
	// The user who uploaded this custom icon.
	CreatedBy *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who uploaded this custom icon."`

	// A timestamp when this custom icon was uploaded. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this custom icon was uploaded. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*CustomIcon) TableName() string {
	return "custom_icons"
}

// ReadAll returns all custom icons in the library, optionally filtered by name.
// The library is shared instance-wide: any authenticated user may search and use it.
func (ci *CustomIcon) ReadAll(s *xorm.Session, _ web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	query := s.Where("1 = 1")
	if search != "" {
		query = query.And("name LIKE ?", "%"+search+"%")
	}

	icons := []*CustomIcon{}
	err = query.
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&icons)
	if err != nil {
		return nil, 0, 0, err
	}

	countQuery := s.Where("1 = 1")
	if search != "" {
		countQuery = countQuery.And("name LIKE ?", "%"+search+"%")
	}
	total, err := countQuery.Count(&CustomIcon{})
	if err != nil {
		return nil, 0, 0, err
	}

	userIDs := make([]int64, 0, len(icons))
	for _, ci := range icons {
		userIDs = append(userIDs, ci.CreatedByID)
	}
	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	for _, i := range icons {
		if createdBy, has := users[i.CreatedByID]; has {
			i.CreatedBy = createdBy
		}
	}

	return icons, len(icons), total, nil
}

// Delete removes a custom icon and its underlying file. Only the uploader or an
// instance admin may delete it (enforced in CanDelete); links still pointing at
// its file keep rendering their icon since files are only deleted once unused.
func (ci *CustomIcon) Delete(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ?", ci.ID).NoAutoCondition().Get(ci)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCustomIconDoesNotExist{ID: ci.ID}
	}

	if _, err = s.Where("id = ?", ci.ID).Delete(&CustomIcon{}); err != nil {
		return err
	}

	return deleteCustomIconFileIfUnused(s, ci.FileID)
}

// customIconFileStillReferenced reports whether any project link, task link, or
// library entry still points at fileID. A file can be shared by many links once
// they attach it from the library, so callers must check this before deleting a
// file out from under something else that still references it — never delete
// unconditionally just because "this one link" stopped using it.
func customIconFileStillReferenced(s *xorm.Session, fileID int64) (bool, error) {
	count, err := s.Where("custom_icon_id = ?", fileID).Count(&ProjectLink{})
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	count, err = s.Where("custom_icon_id = ?", fileID).Count(&TaskLink{})
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	count, err = s.Where("file_id = ?", fileID).Count(&CustomIcon{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// deleteCustomIconFileIfUnused deletes fileID's underlying file, but only if
// nothing else (another link, or a library entry) still references it. Call this
// instead of deleting a replaced custom icon's file directly. A no-op for id 0
// (nothing was set).
func deleteCustomIconFileIfUnused(s *xorm.Session, fileID int64) error {
	if fileID == 0 {
		return nil
	}

	stillReferenced, err := customIconFileStillReferenced(s, fileID)
	if err != nil {
		return err
	}
	if stillReferenced {
		return nil
	}

	if err := (&files.File{ID: fileID}).Delete(s); err != nil && !files.IsErrFileDoesNotExist(err) {
		return err
	}
	return nil
}

// CreateCustomIcon uploads f as a new named custom icon in the shared library.
func CreateCustomIcon(s *xorm.Session, name string, f io.ReadSeeker, realname string, realsize uint64, a web.Auth) (*CustomIcon, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	file, err := files.CreateWithSession(s, f, realname, realsize, a)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return nil, ErrCustomIconIsTooLarge{Size: realsize}
		}
		return nil, err
	}

	ci := &CustomIcon{
		Name:        name,
		FileID:      file.ID,
		CreatedByID: u.ID,
		CreatedBy:   u,
	}
	if _, err := s.Insert(ci); err != nil {
		_ = file.Delete(s)
		return nil, err
	}

	return ci, nil
}

// GetCustomIconFile loads the file for a library icon. Any authenticated user may
// read it, matching the library's shared, instance-wide read access.
func GetCustomIconFile(a web.Auth, customIconID int64) (f *files.File, err error) {
	s := db.NewSession()
	defer s.Close()

	ci := &CustomIcon{ID: customIconID}
	can, _, err := ci.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}

	exists, err := s.Where("id = ?", customIconID).Get(ci)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !exists {
		_ = s.Rollback()
		return nil, ErrCustomIconDoesNotExist{ID: customIconID}
	}

	f = &files.File{ID: ci.FileID}
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

// GetCustomIconFileID returns the underlying file id for a library icon, so a
// project/task link can attach it directly (copying the file reference, not the
// library row). Any authenticated user may read it.
func GetCustomIconFileID(s *xorm.Session, a web.Auth, customIconID int64) (int64, error) {
	ci := &CustomIcon{ID: customIconID}
	can, _, err := ci.CanRead(s, a)
	if err != nil {
		return 0, err
	}
	if !can {
		return 0, ErrGenericForbidden{}
	}

	exists, err := s.Where("id = ?", customIconID).Get(ci)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrCustomIconDoesNotExist{ID: customIconID}
	}

	return ci.FileID, nil
}
