package inmemory

import (
	"sync"
	"testing"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
)

// createTestLink creates a test link
func createTestLink(url string, status models.LinkStatus) models.Link {
	return models.Link{
		URL:       url,
		Status:    status,
		Duration:  100 * time.Millisecond,
		CheckedAt: time.Now(),
	}
}

func TestStorage_InsertMany(t *testing.T) {
	t.Run("first insert returns number 1", func(t *testing.T) {
		storage := New()
		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}

		num, err := storage.InsertMany(links)

		if err != nil {
			t.Fatalf("InsertMany() error = %v, want nil", err)
		}
		if num != 1 {
			t.Errorf("InsertMany() num = %d, want 1", num)
		}
	})

	t.Run("sequential inserts return incremental numbers", func(t *testing.T) {
		storage := New()

		links1 := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}
		links2 := []models.Link{
			createTestLink("https://google.com", models.LinkStatusAvailable),
		}
		links3 := []models.Link{
			createTestLink("https://github.com", models.LinkStatusNotAvailable),
		}

		num1, err := storage.InsertMany(links1)
		if err != nil {
			t.Fatalf("InsertMany() first batch error = %v, want nil", err)
		}
		if num1 != 1 {
			t.Errorf("InsertMany() first batch num = %d, want 1", num1)
		}

		num2, err := storage.InsertMany(links2)
		if err != nil {
			t.Fatalf("InsertMany() second batch error = %v, want nil", err)
		}
		if num2 != 2 {
			t.Errorf("InsertMany() second batch num = %d, want 2", num2)
		}

		num3, err := storage.InsertMany(links3)
		if err != nil {
			t.Fatalf("InsertMany() third batch error = %v, want nil", err)
		}
		if num3 != 3 {
			t.Errorf("InsertMany() third batch num = %d, want 3", num3)
		}
	})

	t.Run("empty slice returns error", func(t *testing.T) {
		storage := New()
		links := []models.Link{}

		num, err := storage.InsertMany(links)

		if err == nil {
			t.Error("InsertMany() error = nil, want error")
		}
		if num != 0 {
			t.Errorf("InsertMany() num = %d, want 0 when error occurs", num)
		}
	})

	t.Run("nil slice returns error", func(t *testing.T) {
		storage := New()
		var links []models.Link = nil

		num, err := storage.InsertMany(links)

		if err == nil {
			t.Error("InsertMany() error = nil, want error")
		}
		if num != 0 {
			t.Errorf("InsertMany() num = %d, want 0 when error occurs", num)
		}
	})

	t.Run("large batch returns number and no error", func(t *testing.T) {
		storage := New()
		const size = 1000
		links := make([]models.Link, size)

		for i := 0; i < size; i++ {
			links[i] = createTestLink("https://example.com", models.LinkStatusAvailable)
		}

		num, err := storage.InsertMany(links)

		if err != nil {
			t.Fatalf("InsertMany() error = %v, want nil", err)
		}
		if num != 1 {
			t.Errorf("InsertMany() num = %d, want 1", num)
		}
	})

	t.Run("non-empty slice returns nil error", func(t *testing.T) {
		storage := New()
		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}

		_, err := storage.InsertMany(links)

		if err != nil {
			t.Errorf("InsertMany() error = %v, want nil", err)
		}
	})
}

func TestStorage_InsertMany_Concurrency(t *testing.T) {
	t.Run("concurrent inserts return unique sequential numbers", func(t *testing.T) {
		storage := New()
		const numGoroutines = 50
		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}

		var wg sync.WaitGroup
		results := make(chan int, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				num, err := storage.InsertMany(links)
				if err != nil {
					errors <- err
					return
				}
				results <- num
			}()
		}

		wg.Wait()
		close(results)
		close(errors)

		if len(errors) > 0 {
			t.Fatalf("InsertMany() returned %d errors during concurrent inserts", len(errors))
		}

		numbers := make(map[int]bool)
		for num := range results {
			numbers[num] = true
		}

		if len(numbers) != numGoroutines {
			t.Errorf("InsertMany() returned %d unique numbers, want %d", len(numbers), numGoroutines)
		}

		// Check that numbers are in the range from 1 to numGoroutines
		for i := 1; i <= numGoroutines; i++ {
			if !numbers[i] {
				t.Errorf("InsertMany() missing number %d in results", i)
			}
		}
	})

	t.Run("concurrent inserts with different batch sizes", func(t *testing.T) {
		storage := New()
		const numGoroutines = 10

		var wg sync.WaitGroup
		results := make(chan int, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(batchSize int) {
				defer wg.Done()
				links := make([]models.Link, batchSize)
				for j := 0; j < batchSize; j++ {
					links[j] = createTestLink("https://example.com", models.LinkStatusAvailable)
				}
				num, err := storage.InsertMany(links)
				if err != nil {
					t.Errorf("InsertMany() error = %v, want nil for batch size %d", err, batchSize)
					return
				}
				results <- num
			}(i + 1)
		}

		wg.Wait()
		close(results)

		numbers := make(map[int]bool)
		for num := range results {
			numbers[num] = true
		}

		if len(numbers) != numGoroutines {
			t.Errorf("InsertMany() returned %d unique numbers, want %d", len(numbers), numGoroutines)
		}
	})
}
