package link

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
)

// mockRepository is a mock implementation of linkRepository interface.
type mockRepository struct {
	insertManyFunc func(links []models.Link) (int, error)
	getByNumsFunc  func(linksNum []int) ([]models.Links, error)
	getAllFunc     func() ([]models.Links, error)
}

func (m *mockRepository) InsertMany(links []models.Link) (int, error) {
	if m.insertManyFunc != nil {
		return m.insertManyFunc(links)
	}
	return 1, nil
}

func (m *mockRepository) GetByNums(linksNum []int) ([]models.Links, error) {
	if m.getByNumsFunc != nil {
		return m.getByNumsFunc(linksNum)
	}
	return []models.Links{}, nil
}

func (m *mockRepository) GetAll() ([]models.Links, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc()
	}
	return []models.Links{}, nil
}

// mockURLChecker is a mock implementation of urlChecker interface.
type mockURLChecker struct {
	checkFunc func(ctx context.Context, url string) models.Link
}

func (m *mockURLChecker) CheckURLWithContext(ctx context.Context, url string) models.Link {
	if m.checkFunc != nil {
		return m.checkFunc(ctx, url)
	}
	return models.Link{
		URL:       url,
		Status:    models.LinkStatusAvailable,
		Duration:  100 * time.Millisecond,
		CheckedAt: time.Now(),
	}
}

// mockPDFGenerator is a mock implementation of PDF generator.
type mockPDFGenerator struct {
	generateFunc func(linksSlice []models.Links) (*bytes.Buffer, error)
}

func (m *mockPDFGenerator) GenerateMultipleReports(linksSlice []models.Links) (*bytes.Buffer, error) {
	if m.generateFunc != nil {
		return m.generateFunc(linksSlice)
	}
	return bytes.NewBufferString("mock pdf content"), nil
}

// createTestLink creates a test link for convenience.
func createTestLink(url string, status models.LinkStatus) models.Link {
	return models.Link{
		URL:       url,
		Status:    status,
		Duration:  100 * time.Millisecond,
		CheckedAt: time.Now(),
	}
}

func TestService_GetAll(t *testing.T) {
	t.Run("successful get all", func(t *testing.T) {
		expectedLinks := []models.Links{
			{
				LinksNum: 1,
				Links: []models.Link{
					createTestLink("https://example.com", models.LinkStatusAvailable),
				},
			},
			{
				LinksNum: 2,
				Links: []models.Link{
					createTestLink("https://google.com", models.LinkStatusAvailable),
				},
			},
		}

		repo := &mockRepository{
			getAllFunc: func() ([]models.Links, error) {
				return expectedLinks, nil
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.GetAll(ctx)

		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if len(result) != len(expectedLinks) {
			t.Fatalf("GetAll() returned %d groups, want %d", len(result), len(expectedLinks))
		}
		if result[0].LinksNum != expectedLinks[0].LinksNum {
			t.Errorf("GetAll() first group LinksNum = %d, want %d", result[0].LinksNum, expectedLinks[0].LinksNum)
		}
	})

	t.Run("returns empty slice when no links", func(t *testing.T) {
		repo := &mockRepository{
			getAllFunc: func() ([]models.Links, error) {
				return []models.Links{}, nil
			},
		}

		service := &Service{
			repository:   repo,
			urlChecker:   &mockURLChecker{},
			pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
			workerCount:  2,
		}

		ctx := context.Background()
		result, err := service.GetAll(ctx)

		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if len(result) != 0 {
			t.Errorf("GetAll() returned %d groups, want 0", len(result))
		}
	})

	t.Run("handles repository error", func(t *testing.T) {
		repo := &mockRepository{
			getAllFunc: func() ([]models.Links, error) {
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
		_, err := service.GetAll(ctx)

		if err == nil {
			t.Error("GetAll() error = nil, want error")
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

		_, err := service.GetAll(ctx)

		if err == nil {
			t.Error("GetAll() error = nil, want context.Canceled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("GetAll() error = %v, want context.Canceled", err)
		}
	})
}
