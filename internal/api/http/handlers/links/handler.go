package links

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
)

type CheckLinksRequest struct {
	Links []string `json:"links"`
}

type CheckLinksResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

type service interface {
	CheckMany(ctx context.Context, links []string) (models.LinksResponse, error)
	GenerateReport(ctx context.Context, links_num []int) (*bytes.Buffer, error)
	GetAll(ctx context.Context) ([]models.Links, error)
}

type Handler struct {
	service service
}

func New(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, _ = context.WithTimeout(ctx, 3*time.Second)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if len(req.Links) == 0 {
		http.Error(w, "Links array cannot be empty", http.StatusBadRequest)
		return
	}

	result, err := h.service.CheckMany(ctx, req.Links)

	if err != nil {
		if err == context.DeadlineExceeded {
			http.Error(w, "Link check timeout", http.StatusRequestTimeout)
			return
		}
		if err == context.Canceled {
			http.Error(w, "Request canceled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.LinksNum) == 0 {
		http.Error(w, "Links_num array cannot be empty", http.StatusBadRequest)
		return
	}

	pdfBuffer, err := h.service.GenerateReport(ctx, req.LinksNum)
	if err != nil {
		http.Error(w, "Failed to generate report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, хочет ли клиент JSON ответ или PDF
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "application/json") {
		// Возвращаем JSON с информацией об отчете
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.GenerateReportResponse{
			Message: "PDF report generated successfully",
			Size:    pdfBuffer.Len(),
		})
		return
	}

	// По умолчанию возвращаем PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=link_report.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	_, err = pdfBuffer.WriteTo(w)
	if err != nil {
		http.Error(w, "Failed to send PDF", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, _ = context.WithTimeout(ctx, 3*time.Second)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.service.GetAll(ctx)

	if err != nil {
		if err == context.DeadlineExceeded {
			http.Error(w, "Get all timeout", http.StatusRequestTimeout)
			return
		}
		if err == context.Canceled {
			http.Error(w, "Request canceled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
