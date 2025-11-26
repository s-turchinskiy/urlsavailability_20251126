// Package models Модели
package models

//go:generate easyjson -all models.go

type URLsRequest struct {
	Links []string `json:"links" validate:"required"`
}

type URLsResponse struct {
	Links map[string]string `json:"links" validate:"required"`
	Num   uint64            `json:"links_num" validate:"required"`
}

type PDFRequest struct {
	Nums []uint64 `json:"links_list" validate:"required"`
}

type URLsKit map[string]bool

type Data map[uint64]URLsKit
type FileStore struct {
	Data Data
	Num  uint64
}

const (
	StringAvailable    = "available"
	StringNotAvailable = "not available"
)

func (k URLsKit) ConvertToReadableView() map[string]string {

	result := make(map[string]string, len(k))
	for idx, val := range k {
		if val == true {
			result[idx] = StringAvailable
		} else {
			result[idx] = StringNotAvailable
		}
	}

	return result
}
