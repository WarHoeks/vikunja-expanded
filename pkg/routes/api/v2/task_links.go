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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	webfiles "code.vikunja.io/api/pkg/web/files"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
)

// models.TaskLink.ReadAll returns []*models.TaskLink, so that's the element type.
type taskLinkListBody struct {
	Body Paginated[*models.TaskLink]
}

// RegisterTaskLinkRoutes wires the nested task-links CRUD onto the Huma API.
// Mirrors project_links.go's shape (no ReadOne, so update is PUT only).
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
		Description: "Creates a reference link on the given task. The parent task is taken from the URL, not the body. Requires write access to the task. Set icon to a simple-icons (https://simple-icons.org) slug; leave it empty if you intend to attach a custom icon via the icon upload endpoint instead.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/links",
		Tags:        tags,
	}, taskLinksCreate)

	Register(api, huma.Operation{
		OperationID: "task-links-update",
		Summary:     "Update a task link",
		Description: "Replaces a link's url, title and icon. The link must belong to the task in the path, and write access to that task is required.",
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

	Register(api, huma.Operation{
		OperationID: "task-links-icon-upload",
		Summary:     "Upload a custom icon for a task link",
		Description: "Uploads an image as the link's icon, replacing any simple-icons slug or previously uploaded custom icon. Requires write access to the task. The max size is the server's configured file size limit.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/links/{link}/icon",
		Tags:        tags,
		// +2 MB mirrors Echo's global BodyLimit overhead, same as task-attachments-upload.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, taskLinksIconUpload)

	Register(api, huma.Operation{
		OperationID: "task-links-icon-download",
		Summary:     "Download a task link's custom icon",
		Description: "Returns the raw bytes of a link's uploaded custom icon. Requires read access to the task. 404s if the link has no custom icon (e.g. it uses a simple-icons slug instead).",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/links/{link}/icon",
		Tags:        tags,
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The icon file bytes. The Content-Type header carries the file's mime type.",
				Content: map[string]*huma.MediaType{
					"application/octet-stream": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, taskLinksIconDownload)
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

type taskLinkIconUploadInput struct {
	TaskID  int64 `path:"task" doc:"The id of the task the link belongs to."`
	LinkID  int64 `path:"link" doc:"The id of the link to set the icon on."`
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"icon" required:"true" doc:"The image to use as the link's icon."`
	}]
}

// taskLinksIconUpload is a custom (non-CRUDable) action, so permission
// enforcement and session management are the handler's responsibility here —
// there is no handler.Do* for a multipart upload (see the api-v2-routes skill).
func taskLinksIconUpload(ctx context.Context, in *taskLinkIconUploadInput) (*singleBody[models.TaskLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	tl := &models.TaskLink{ID: in.LinkID, TaskID: in.TaskID}
	can, err := tl.CanUpdate(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	file := in.RawBody.Data().File
	if err := tl.SetCustomIcon(s, file, file.Filename, uint64(file.Size), a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &singleBody[models.TaskLink]{Body: tl}, nil
}

// taskLinksIconDownload exists because no handler.Do* fits a file body; bytes
// stream from the StreamResponse callback without buffering.
func taskLinksIconDownload(ctx context.Context, in *struct {
	TaskID int64 `path:"task" doc:"The id of the task the link belongs to."`
	LinkID int64 `path:"link" doc:"The id of the link whose custom icon to download."`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	f, err := models.GetTaskLinkIconFile(a, in.TaskID, in.LinkID)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		c := humaecho.Unwrap(hctx)
		webfiles.WriteFileDownload((*c).Response(), (*c).Request(), f)
	}}, nil
}
