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

// models.ProjectLink.ReadAll returns []*models.ProjectLink, so that's the element type.
type projectLinkListBody struct {
	Body Paginated[*models.ProjectLink]
}

// RegisterProjectLinkRoutes wires the nested project-links CRUD onto the Huma API.
// Mirrors webhooks.go's shape: no ReadOne (the list is small and fetched whole for
// the sidebar), so AutoPatch synthesises no PATCH here and update is PUT only.
func RegisterProjectLinkRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "project-links-list",
		Summary:     "List a project's links",
		Description: "Returns the reference links (repo, environment, docs, ...) attached to the given project, paginated. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/links",
		Tags:        tags,
	}, projectLinksList)

	Register(api, huma.Operation{
		OperationID: "project-links-create",
		Summary:     "Add a link to a project",
		Description: "Creates a reference link on the given project. The parent project is taken from the URL, not the body. Requires write access to the project. Set icon to a simple-icons (https://simple-icons.org) slug; leave it empty if you intend to attach a custom icon via the icon upload endpoint instead.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/links",
		Tags:        tags,
	}, projectLinksCreate)

	Register(api, huma.Operation{
		OperationID: "project-links-update",
		Summary:     "Update a project link",
		Description: "Replaces a link's url, title and icon. The link must belong to the project in the path, and write access to that project is required.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/links/{link}",
		Tags:        tags,
	}, projectLinksUpdate)

	Register(api, huma.Operation{
		OperationID: "project-links-delete",
		Summary:     "Delete a project link",
		Description: "Deletes a link. The link must belong to the project in the path, and write access to that project is required.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/links/{link}",
		Tags:        tags,
	}, projectLinksDelete)

	Register(api, huma.Operation{
		OperationID: "project-links-icon-upload",
		Summary:     "Upload a custom icon for a project link",
		Description: "Uploads an image as the link's icon, replacing any simple-icons slug or previously uploaded custom icon. Requires write access to the project. The max size is the server's configured file size limit.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/links/{link}/icon",
		Tags:        tags,
		// +2 MB mirrors Echo's global BodyLimit overhead, same as task-attachments-upload.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, projectLinksIconUpload)

	Register(api, huma.Operation{
		OperationID: "project-links-icon-download",
		Summary:     "Download a project link's custom icon",
		Description: "Returns the raw bytes of a link's uploaded custom icon. Requires read access to the project. 404s if the link has no custom icon (e.g. it uses a simple-icons slug instead).",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/links/{link}/icon",
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
	}, projectLinksIconDownload)
}

func init() { AddRouteRegistrar(RegisterProjectLinkRoutes) }

func projectLinksList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*projectLinkListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectLink{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.ProjectLink)
	if !ok {
		return nil, fmt.Errorf("projectLinks.ReadAll returned unexpected type %T (expected []*models.ProjectLink)", result)
	}
	return &projectLinkListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func projectLinksCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.ProjectLink
}) (*singleBody[models.ProjectLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectLink]{Body: &in.Body}, nil
}

func projectLinksUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"link"`
	Body      models.ProjectLink
}) (*singleBody[models.ProjectLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID
	in.Body.ProjectID = in.ProjectID
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectLink]{Body: &in.Body}, nil
}

func projectLinksDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"link"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.ProjectLink{ID: in.ID, ProjectID: in.ProjectID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

type projectLinkIconUploadInput struct {
	ProjectID int64 `path:"project" doc:"The id of the project the link belongs to."`
	LinkID    int64 `path:"link" doc:"The id of the link to set the icon on."`
	RawBody   huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"icon" required:"true" doc:"The image to use as the link's icon."`
	}]
}

// projectLinksIconUpload is a custom (non-CRUDable) action, so permission
// enforcement and session management are the handler's responsibility here —
// there is no handler.Do* for a multipart upload (see the api-v2-routes skill).
func projectLinksIconUpload(ctx context.Context, in *projectLinkIconUploadInput) (*singleBody[models.ProjectLink], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	pl := &models.ProjectLink{ID: in.LinkID, ProjectID: in.ProjectID}
	can, err := pl.CanUpdate(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	file := in.RawBody.Data().File
	if err := pl.SetCustomIcon(s, file, file.Filename, uint64(file.Size), a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &singleBody[models.ProjectLink]{Body: pl}, nil
}

// projectLinksIconDownload exists because no handler.Do* fits a file body; bytes
// stream from the StreamResponse callback without buffering.
func projectLinksIconDownload(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" doc:"The id of the project the link belongs to."`
	LinkID    int64 `path:"link" doc:"The id of the link whose custom icon to download."`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	f, err := models.GetProjectLinkIconFile(a, in.ProjectID, in.LinkID)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		c := humaecho.Unwrap(hctx)
		webfiles.WriteFileDownload((*c).Response(), (*c).Request(), f)
	}}, nil
}
