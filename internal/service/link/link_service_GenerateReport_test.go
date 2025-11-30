package link

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/polonkoevv/linkchecker/internal/models"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
)

func TestService_GenerateReport(t *testing.T) {
	t.Run("successful report generation", func(t *testing.T) {
		links := []models.Links{
			{
				LinksNum: 1,
				Links: []models.Link{
					createTestLink("https://example.com", models.LinkStatusAvailable),
				},
			},
		}

		repo := &mockRepository{
			getByNumsFunc: func(linksNum []int) ([]models.Links, error) {
				return links, nil
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.GenerateReport(ctx, []int{1})

		if err != nil {
			t.Fatalf("GenerateReport() error = %v, want nil", err)
		}
		if result == nil {
			t.Error("GenerateReport() result = nil, want buffer")
		}
		if result.Len() == 0 {
			t.Error("GenerateReport() returned empty buffer")
		}
	})

	t.Run("handles repository error", func(t *testing.T) {
		repo := &mockRepository{
			getByNumsFunc: func(linksNum []int) ([]models.Links, error) {
				return nil, errors.New("repository error")
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		_, err := service.GenerateReport(ctx, []int{1})

		if err == nil {
			t.Error("GenerateReport() error = nil, want error")
		}
	})

	t.Run("handles PDF generator error", func(t *testing.T) {
		links := []models.Links{
			{
				LinksNum: 1,
				Links: []models.Link{
					createTestLink("https://example.com", models.LinkStatusAvailable),
				},
			},
		}

		repo := &mockRepository{
			getByNumsFunc: func(linksNum []int) ([]models.Links, error) {
				return links, nil
			},
		}

		pdfGen := &mockPDFGenerator{
			generateFunc: func(linksSlice []models.Links) (*bytes.Buffer, error) {
				return nil, errors.New("PDF generation error")
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfGen,
			workerCount:  2,
		}

		ctx := context.Background()
		_, err := service.GenerateReport(ctx, []int{1})

		if err == nil {
			t.Error("GenerateReport() error = nil, want error")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		service := &Service{
			repository:   &mockRepository{},
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := service.GenerateReport(ctx, []int{1})

		if err == nil {
			t.Error("GenerateReport() error = nil, want context.Canceled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("GenerateReport() error = %v, want context.Canceled", err)
		}
	})
}
