package links

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
)

// CheckLinksRequest represents a request payload for checking multiple links.
type CheckLinksRequest struct {
	Links []string `json:"links"`
}

type service interface {
	CheckMany(ctx context.Context, links []string) (models.LinksResponse, error)
	GenerateReport(ctx context.Context, linksNum []int) (*bytes.Buffer, error)
	GetAll(ctx context.Context) ([]models.Links, error)
}

// Handler provides HTTP handlers for link checking and reporting.
type Handler struct {
	Service        service
	RequestTimeout time.Duration
}

// New constructs a new Handler with the given service and per-request timeout.
func New(service service, requestTimeout time.Duration) *Handler {
	return &Handler{
		Service:        service,
		RequestTimeout: requestTimeout,
	}
}

// Check handles POST /links and triggers asynchronous link status checks.
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	slog.Info("incoming request",
		slog.String("handler", "Check"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()

	if r.Method != http.MethodPost {
		slog.Warn("method not allowed",
			slog.String("handler", "Check"),
			slog.String("method", r.Method),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckLinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("failed to decode request body",
			slog.String("handler", "Check"),
			slog.Any("error", err),
		)
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if len(req.Links) == 0 {
		slog.Warn("validation failed: links array is empty", slog.String("handler", "Check"))
		http.Error(w, "Links array cannot be empty", http.StatusBadRequest)
		return
	}

	result, err := h.Service.CheckMany(ctx, req.Links)
	if err != nil {
		if err == context.DeadlineExceeded {
			slog.Warn("check links timeout", slog.String("handler", "Check"))
			http.Error(w, "Link check timeout", http.StatusRequestTimeout)
			return
		}
		if err == context.Canceled {
			slog.Warn("request canceled by client", slog.String("handler", "Check"))
			http.Error(w, "Request canceled", http.StatusRequestTimeout)
			return
		}

		slog.Error("check many failed",
			slog.String("handler", "Check"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Debug("links checked successfully",
		slog.String("handler", "Check"),
		slog.Int("links_count", len(req.Links)),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// GenerateReport handles POST /report and returns a PDF or JSON report.
func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	slog.Info("incoming request",
		slog.String("handler", "GenerateReport"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()

	if r.Method != http.MethodPost {
		slog.Warn("method not allowed",
			slog.String("handler", "GenerateReport"),
			slog.String("method", r.Method),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("failed to decode request body",
			slog.String("handler", "GenerateReport"),
			slog.Any("error", err),
		)
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.LinksNum) == 0 {
		slog.Warn("validation failed: links_num array is empty", slog.String("handler", "GenerateReport"))
		http.Error(w, "Links_num array cannot be empty", http.StatusBadRequest)
		return
	}

	pdfBuffer, err := h.Service.GenerateReport(ctx, req.LinksNum)
	if err != nil {
		slog.Error("failed to generate report",
			slog.String("handler", "GenerateReport"),
			slog.Any("error", err),
		)
		http.Error(w, "Failed to generate report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, хочет ли клиент JSON ответ или PDF
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "application/json") {
		slog.Debug("returning JSON report meta",
			slog.String("handler", "GenerateReport"),
			slog.Int("links_num_count", len(req.LinksNum)),
			slog.Int("size_bytes", pdfBuffer.Len()),
		)

		// Возвращаем JSON с информацией об отчете
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.GenerateReportResponse{
			Message: "PDF report generated successfully",
			Size:    pdfBuffer.Len(),
		})
		return
	}

	// По умолчанию возвращаем PDF
	slog.Debug("returning PDF report",
		slog.String("handler", "GenerateReport"),
		slog.Int("links_num_count", len(req.LinksNum)),
		slog.Int("size_bytes", pdfBuffer.Len()),
	)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=link_report.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	if _, err = pdfBuffer.WriteTo(w); err != nil {
		slog.Error("failed to send PDF to client",
			slog.String("handler", "GenerateReport"),
			slog.Any("error", err),
		)
		http.Error(w, "Failed to send PDF", http.StatusInternalServerError)
		return
	}
}

// GetAll handles GET /links and returns all stored link groups.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	slog.Info("incoming request",
		slog.String("handler", "GetAll"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()

	if r.Method != http.MethodGet {
		slog.Warn("method not allowed",
			slog.String("handler", "GetAll"),
			slog.String("method", r.Method),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.Service.GetAll(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			slog.Warn("get all timeout", slog.String("handler", "GetAll"))
			http.Error(w, "Get all timeout", http.StatusRequestTimeout)
			return
		}
		if err == context.Canceled {
			slog.Warn("request canceled by client", slog.String("handler", "GetAll"))
			http.Error(w, "Request canceled", http.StatusRequestTimeout)
			return
		}

		slog.Error("get all links failed",
			slog.String("handler", "GetAll"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Debug("get all links succeeded",
		slog.String("handler", "GetAll"),
		slog.Int("groups_count", len(result)),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
