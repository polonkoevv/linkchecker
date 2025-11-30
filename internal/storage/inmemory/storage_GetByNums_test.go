package inmemory

import (
	"errors"
	"sync"
	"testing"

	"github.com/polonkoevv/linkchecker/internal/models"
)

func TestStorage_GetByNums(t *testing.T) {
	t.Run("get existing single group", func(t *testing.T) {
		storage := New()
		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}

		num, err := storage.InsertMany(links)
		if err != nil {
			t.Fatalf("InsertMany() error = %v, want nil", err)
		}

		result, err := storage.GetByNums([]int{num})
		if err != nil {
			t.Fatalf("GetByNums() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("GetByNums() returned %d groups, want 1", len(result))
		}
		if result[0].LinksNum != num {
			t.Errorf("GetByNums() LinksNum = %d, want %d", result[0].LinksNum, num)
		}
		if len(result[0].Links) != len(links) {
			t.Errorf("GetByNums() returned %d links, want %d", len(result[0].Links), len(links))
		}
	})

	t.Run("get multiple existing groups", func(t *testing.T) {
		storage := New()

		links1 := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}
		links2 := []models.Link{
			createTestLink("https://google.com", models.LinkStatusAvailable),
		}

		num1, _ := storage.InsertMany(links1)
		num2, _ := storage.InsertMany(links2)

		result, err := storage.GetByNums([]int{num1, num2})
		if err != nil {
			t.Fatalf("GetByNums() error = %v, want nil", err)
		}
		if len(result) != 2 {
			t.Fatalf("GetByNums() returned %d groups, want 2", len(result))
		}
	})

	t.Run("get non-existent group returns error", func(t *testing.T) {
		storage := New()

		result, err := storage.GetByNums([]int{999})

		if err == nil {
			t.Error("GetByNums() error = nil, want error")
		}
		if result != nil {
			t.Errorf("GetByNums() result = %v, want nil", result)
		}
	})

	t.Run("get partial results returns found groups without error", func(t *testing.T) {
		storage := New()

		links1 := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}
		links2 := []models.Link{
			createTestLink("https://google.com", models.LinkStatusAvailable),
		}

		num1, _ := storage.InsertMany(links1)
		num2, _ := storage.InsertMany(links2)

		result, err := storage.GetByNums([]int{num1, 999, num2})

		if err != nil {
			t.Fatalf("GetByNums() error = %v, want nil (partial results should not error)", err)
		}
		if len(result) != 2 {
			t.Fatalf("GetByNums() returned %d groups, want 2", len(result))
		}
	})

	t.Run("concurrent reads return correct data", func(t *testing.T) {
		storage := New()
		const numGroups = 5

		groupNums := make([]int, numGroups)
		for i := 0; i < numGroups; i++ {
			links := []models.Link{
				createTestLink("https://example.com", models.LinkStatusAvailable),
			}
			num, _ := storage.InsertMany(links)
			groupNums[i] = num
		}

		const numReaders = 10
		var wg sync.WaitGroup
		errs := make(chan error, numReaders)

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := storage.GetByNums(groupNums)
				if err != nil {
					errs <- err
					return
				}
				if len(result) != numGroups {
					errs <- errors.New("incorrect number of groups returned")
					return
				}
			}()
		}

		wg.Wait()
		close(errs)

		if len(errs) > 0 {
			t.Fatalf("GetByNums() returned %d errors during concurrent reads", len(errs))
		}
	})
}
