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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanRead: the custom icon library is shared instance-wide, readable by any authenticated user.
func (ci *CustomIcon) CanRead(_ *xorm.Session, a web.Auth) (bool, int, error) {
	_, isLinkShare := a.(*LinkSharing)
	if isLinkShare {
		return false, 0, nil
	}
	return true, int(PermissionRead), nil
}

// CanCreate: any authenticated user may add a new icon to the shared library.
func (ci *CustomIcon) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	_, isLinkShare := a.(*LinkSharing)
	return !isLinkShare, nil
}

// CanDelete: only the uploader or an instance admin may remove a library icon.
func (ci *CustomIcon) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}

	existing := &CustomIcon{ID: ci.ID}
	has, err := s.Get(existing)
	if err != nil {
		return false, err
	}
	if !has {
		return false, nil
	}

	return existing.CreatedByID == a.GetID(), nil
}
