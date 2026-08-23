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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// models.TaskLink.ReadAll returns []*models.TaskLink, so that's the element type.
type taskLinkListBody struct {
	Body Paginated[*models.TaskLink]
}

// RegisterTaskLinkRoutes wires the nested task-links CRUD onto the Huma API.
// Mirrors project_links.go's shape (no ReadOne, so update is PUT only). Unlike
// project links, task links have no icon — just a URL and a title.
func RegisterTaskLinkRoutes(api huma.API) {
	tags := []string{"task"}

	Register(api, huma.Operation{
		OperationID: "task-links-list",
		Summary:     "List a task's links",
		Description: "Returns the reference links (repo, environment, docs, ...) attached to the given task, paginated. Requires read access to the task.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/links",
		Tags:        tags,
	}, taskLinksList)

	Register(api, huma.Operation{
		OperationID: "task-links-create",
		Summary:     "Add a link to a task",
		Description: "Creates a reference link on the given task. The parent task is taken from the URL, not the body. Requires write access to the task.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/links",
		Tags:        tags,
	}, taskLinksCreate)

	Register(api, huma.Operation{
		OperationID: "task-links-update",
		Summary:     "Update a task link",
		Description: "Replaces a link's url and title. The link must belong to the task in the path, and write access to that task is required.",
		Method:      http.MethodPut,
		Path:        "/tasks/{task}/links/{link}",
		Tags:        tags,
	}, taskLinksUpdate)

	Register(api, huma.Operation{
		OperationID: "task-links-delete",
		Summary:     "Delete a task link",
		Description: "Deletes a link. The link must belong to the task in the path, and write access to that task is required.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/links/{link}",
		Tags:        tags,
	}, taskLinksDelete)
}

func init() { AddRouteRegistrar(RegisterTaskLinkRoutes) }

func taskLinksList(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ListParams
}) (*taskLinkListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TaskLink{TaskID: in.TaskID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.TaskLink)
	if !ok {
		return nil, fmt.Errorf("taskLinks.ReadAll returned unexpected type %T (expected []*models.TaskLink)", result)
	}
	return &taskLinkListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func taskLinksCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	Body   models.TaskLink
}) (*singleBody[models.TaskLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskLink]{Body: &in.Body}, nil
}

func taskLinksUpdate(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ID     int64 `path:"link"`
	Body   models.TaskLink
}) (*singleBody[models.TaskLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID
	in.Body.TaskID = in.TaskID
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskLink]{Body: &in.Body}, nil
}

func taskLinksDelete(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ID     int64 `path:"link"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TaskLink{ID: in.ID, TaskID: in.TaskID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
