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

// models.CustomIcon.ReadAll returns []*models.CustomIcon, so that's the element type.
type customIconListBody struct {
	Body Paginated[*models.CustomIcon]
}

// RegisterCustomIconRoutes wires the shared, instance-wide custom icon library onto
// the Huma API: search/list, upload, delete, and download the underlying image.
func RegisterCustomIconRoutes(api huma.API) {
	tags := []string{"custom-icons"}

	Register(api, huma.Operation{
		OperationID: "custom-icons-list",
		Summary:     "Search the custom icon library",
		Description: "Returns custom icons uploaded to the shared, instance-wide library, optionally filtered by name (q). Any authenticated user can search and use them.",
		Method:      http.MethodGet,
		Path:        "/custom-icons",
		Tags:        tags,
	}, customIconsList)

	Register(api, huma.Operation{
		OperationID: "custom-icons-upload",
		Summary:     "Upload a new custom icon to the library",
		Description: "Uploads an image as a new, named entry in the shared custom icon library, reusable across any project or task link. Requires only an authenticated (non-link-share) user.",
		Method:      http.MethodPost,
		Path:        "/custom-icons",
		Tags:        tags,
		// +2 MB mirrors Echo's global BodyLimit overhead, same as task-attachments-upload.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, customIconsUpload)

	Register(api, huma.Operation{
		OperationID: "custom-icons-delete",
		Summary:     "Delete a custom icon from the library",
		Description: "Deletes a library entry. Only the user who uploaded it, or an instance admin, may delete it. Links that already reference its image keep rendering it.",
		Method:      http.MethodDelete,
		Path:        "/custom-icons/{customicon}",
		Tags:        tags,
	}, customIconsDelete)

	Register(api, huma.Operation{
		OperationID: "custom-icons-download",
		Summary:     "Download a custom icon's image",
		Description: "Returns the raw bytes of a library icon's image. Any authenticated user can read it.",
		Method:      http.MethodGet,
		Path:        "/custom-icons/{customicon}/icon",
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
	}, customIconsDownload)
}

func init() { AddRouteRegistrar(RegisterCustomIconRoutes) }

func customIconsList(ctx context.Context, in *struct {
	ListParams
}) (*customIconListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.CustomIcon{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.CustomIcon)
	if !ok {
		return nil, fmt.Errorf("customIcons.ReadAll returned unexpected type %T (expected []*models.CustomIcon)", result)
	}
	return &customIconListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

type customIconUploadInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Name string        `form:"name" required:"true" doc:"The searchable name for this custom icon."`
		File huma.FormFile `form:"file" required:"true" doc:"The image to add to the custom icon library."`
	}]
}

// customIconsUpload is a custom (non-CRUDable) action: creation is a combined
// name+file multipart upload, which has no handler.Do* equivalent (see the
// api-v2-routes skill's "Non-CRUDable / custom routes" section). Permission
// enforcement (CanCreate) happens inside models.CreateCustomIcon indirectly via
// the same open-to-any-authenticated-user rule as CanCreate; there is no
// project/task scope to check here, so authFromCtx is the only gate needed.
func customIconsUpload(ctx context.Context, in *customIconUploadInput) (*singleBody[models.CustomIcon], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	data := in.RawBody.Data()
	icon, err := models.CreateCustomIcon(s, data.Name, data.File, data.File.Filename, uint64(data.File.Size), a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &singleBody[models.CustomIcon]{Body: icon}, nil
}

func customIconsDelete(ctx context.Context, in *struct {
	ID int64 `path:"customicon"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.CustomIcon{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// customIconsDownload exists because no handler.Do* fits a file body; bytes
// stream from the StreamResponse callback without buffering.
func customIconsDownload(ctx context.Context, in *struct {
	ID int64 `path:"customicon" doc:"The id of the custom icon whose image to download."`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	f, err := models.GetCustomIconFile(a, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		c := humaecho.Unwrap(hctx)
		webfiles.WriteFileDownload((*c).Response(), (*c).Request(), f)
	}}, nil
}
