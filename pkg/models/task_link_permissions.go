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

// CanRead checks if a user has read access to a task link (delegates to the parent task)
func (tl *TaskLink) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	t := &Task{ID: tl.TaskID}
	return t.CanRead(s, a)
}

// CanCreate checks if a user can add a link to a task
func (tl *TaskLink) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	t, err := GetTaskByIDSimple(s, tl.TaskID)
	if err != nil {
		return false, err
	}
	return t.CanCreate(s, a)
}

// CanUpdate checks if a user can update a task link
func (tl *TaskLink) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: tl.TaskID}
	return t.CanWrite(s, a)
}

// CanDelete checks if a user can delete a task link
func (tl *TaskLink) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: tl.TaskID}
	return t.CanWrite(s, a)
}
