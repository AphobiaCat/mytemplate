package common

import "mime/multipart"

type UploadCsvRequest struct {
	File *multipart.FileHeader `form:"file" validate:"required"`
}

type UploadCsvResponse struct {
	Count int64 `json:"count"`
}
