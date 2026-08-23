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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projectLinks20260823115017 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID    int64     `xorm:"bigint not null index"`
	URL          string    `xorm:"varchar(2000) not null"`
	Title        string    `xorm:"varchar(250) not null"`
	Icon         string    `xorm:"varchar(100) null"`
	CustomIconID int64     `xorm:"bigint null"`
	CreatedByID  int64     `xorm:"bigint not null"`
	Created      time.Time `xorm:"created not null"`
	Updated      time.Time `xorm:"updated not null"`
}

func (projectLinks20260823115017) TableName() string {
	return "project_links"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260823115017",
		Description: "Add project_links table (vikunja-expanded: project web links)",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(projectLinks20260823115017{}) //nolint:forbidigo // brand-new table, nothing to drop
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(projectLinks20260823115017{})
		},
	})
}
