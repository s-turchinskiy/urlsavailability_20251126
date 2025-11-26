package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/s-turchinskiy/urlsavailability/internal/service"
	"github.com/s-turchinskiy/urlsavailability/models"

	"github.com/mailru/easyjson"
)

type HTTPServerHandlers struct {
	Service service.Servicer
}

const (
	ContentTypeTextHTML         = "text/html; charset=utf-8"
	ContentTypeTextPlain        = "text/plain"
	ContentTypeTextPlainCharset = "text/plain; charset=utf-8"
	ContentTypeApplicationJSON  = "application/json"
	ContentTypeApplicationPDF   = "application/pdf"
)

func NewHandlers(
	service service.Servicer) *HTTPServerHandlers {
	return &HTTPServerHandlers{
		Service: service,
	}

}

func (h *HTTPServerHandlers) Statuses(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var req models.URLsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		SendBadRequestWithError(w, err)
		return
	}

	result, num, err := h.Service.Availability(r.Context(), req.Links)
	if err != nil {
		SendBadRequestWithError(w, err)
		return
	}

	resp := models.URLsResponse{Links: result.ConvertToReadableView(), Num: num}
	rawBytes, err := easyjson.Marshal(resp)
	if err != nil {
		SendBadRequestWithError(w, err)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(rawBytes)

}

func (h *HTTPServerHandlers) PDF(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var req models.PDFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		SendBadRequestWithError(w, err)
		return
	}

	bytes, err := h.Service.GetPDF(r.Context(), req.Nums)
	if err != nil {
		SendBadRequestWithError(w, err)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationPDF)
	w.Write(bytes)

}

func SendBadRequestWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", ContentTypeTextHTML)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))

}
