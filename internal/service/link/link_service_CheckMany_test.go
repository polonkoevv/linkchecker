package link

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
)

func TestService_CheckMany(t *testing.T) {
	t.Run("successful check with single link", func(t *testing.T) {
		repo := &mockRepository{
			insertManyFunc: func(links []models.Link) (int, error) {
				if len(links) != 1 {
					t.Errorf("InsertMany() called with %d links, want 1", len(links))
				}
				return 1, nil
			},
		}

		checker := &mockURLChecker{
			checkFunc: func(ctx context.Context, url string) models.Link {
				return createTestLink(url, models.LinkStatusAvailable)
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   checker,
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.CheckMany(ctx, []string{"https://example.com"})

		if err != nil {
			t.Fatalf("CheckMany() error = %v, want nil", err)
		}
		if result.LinksNum != 1 {
			t.Errorf("CheckMany() LinksNum = %d, want 1", result.LinksNum)
		}
		if len(result.Links) != 1 {
			t.Errorf("CheckMany() returned %d links, want 1", len(result.Links))
		}
		if result.Links["https://example.com"] != models.LinkStatusAvailable {
			t.Errorf("CheckMany() link status = %s, want %s", result.Links["https://example.com"], models.LinkStatusAvailable)
		}
	})

	t.Run("deduplicates links", func(t *testing.T) {
		repo := &mockRepository{
			insertManyFunc: func(links []models.Link) (int, error) {
				if len(links) != 2 {
					t.Errorf("InsertMany() called with %d links, want 2 (duplicates removed)", len(links))
				}
				return 1, nil
			},
		}

		checker := &mockURLChecker{
			checkFunc: func(ctx context.Context, url string) models.Link {
				return createTestLink(url, models.LinkStatusAvailable)
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   checker,
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.CheckMany(ctx, []string{
			"https://example.com",
			"https://example.com", // duplicate
			"https://google.com",
		})

		if err != nil {
			t.Fatalf("CheckMany() error = %v, want nil", err)
		}
		if len(result.Links) != 2 {
			t.Errorf("CheckMany() returned %d links, want 2", len(result.Links))
		}
	})

	t.Run("returns empty response for empty links", func(t *testing.T) {
		service := &Service{
			repository:   &mockRepository{},
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.CheckMany(ctx, []string{})

		if err != nil {
			t.Fatalf("CheckMany() error = %v, want nil", err)
		}
		if result.LinksNum != 0 {
			t.Errorf("CheckMany() LinksNum = %d, want 0", result.LinksNum)
		}
		if len(result.Links) != 0 {
			t.Errorf("CheckMany() returned %d links, want 0", len(result.Links))
		}
	})

	t.Run("handles repository error", func(t *testing.T) {
		repo := &mockRepository{
			insertManyFunc: func(links []models.Link) (int, error) {
				return 0, errors.New("repository error")
			},
		}

		checker := &mockURLChecker{
			checkFunc: func(ctx context.Context, url string) models.Link {
				return createTestLink(url, models.LinkStatusAvailable)
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   checker,
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		_, err := service.CheckMany(ctx, []string{"https://example.com"})

		if err == nil {
			t.Error("CheckMany() error = nil, want error")
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
		cancel() // Cancel immediately

		_, err := service.CheckMany(ctx, []string{"https://example.com"})

		if err == nil {
			t.Error("CheckMany() error = nil, want context.Canceled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("CheckMany() error = %v, want context.Canceled", err)
		}
	})

	t.Run("handles context timeout", func(t *testing.T) {
		service := &Service{
			repository:   &mockRepository{},
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := service.CheckMany(ctx, []string{"https://example.com"})

		if err == nil {
			t.Error("CheckMany() error = nil, want context.DeadlineExceeded")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("CheckMany() error = %v, want context.DeadlineExceeded", err)
		}
	})
}
