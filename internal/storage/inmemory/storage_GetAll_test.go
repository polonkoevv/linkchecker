package inmemory

import (
	"sync"
	"testing"

	"github.com/polonkoevv/linkchecker/internal/models"
)

func TestStorage_GetAll(t *testing.T) {
	t.Run("get all from empty storage returns empty slice", func(t *testing.T) {
		storage := New()

		result, err := storage.GetAll()

		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if result == nil {
			t.Error("GetAll() result = nil, want empty slice")
		}
		if len(result) != 0 {
			t.Errorf("GetAll() returned %d groups, want 0", len(result))
		}
	})

	t.Run("get all returns single group", func(t *testing.T) {
		storage := New()
		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}

		num, err := storage.InsertMany(links)
		if err != nil {
			t.Fatalf("InsertMany() error = %v, want nil", err)
		}

		result, err := storage.GetAll()
		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("GetAll() returned %d groups, want 1", len(result))
		}
		if result[0].LinksNum != num {
			t.Errorf("GetAll() LinksNum = %d, want %d", result[0].LinksNum, num)
		}
		if len(result[0].Links) != len(links) {
			t.Errorf("GetAll() returned %d links, want %d", len(result[0].Links), len(links))
		}
	})

	t.Run("get all returns multiple groups", func(t *testing.T) {
		storage := New()

		links1 := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
		}
		links2 := []models.Link{
			createTestLink("https://google.com", models.LinkStatusAvailable),
			createTestLink("https://github.com", models.LinkStatusNotAvailable),
		}
		links3 := []models.Link{
			createTestLink("https://stackoverflow.com", models.LinkStatusAvailable),
		}

		num1, _ := storage.InsertMany(links1)
		num2, _ := storage.InsertMany(links2)
		num3, _ := storage.InsertMany(links3)

		result, err := storage.GetAll()
		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if len(result) != 3 {
			t.Fatalf("GetAll() returned %d groups, want 3", len(result))
		}

		// Проверяем, что все группы присутствуют
		foundNums := make(map[int]bool)
		for _, group := range result {
			foundNums[group.LinksNum] = true
		}

		if !foundNums[num1] {
			t.Errorf("GetAll() missing group %d", num1)
		}
		if !foundNums[num2] {
			t.Errorf("GetAll() missing group %d", num2)
		}
		if !foundNums[num3] {
			t.Errorf("GetAll() missing group %d", num3)
		}
	})

	t.Run("get all preserves links data", func(t *testing.T) {
		storage := New()

		links := []models.Link{
			createTestLink("https://example.com", models.LinkStatusAvailable),
			createTestLink("https://google.com", models.LinkStatusNotAvailable),
		}

		storage.InsertMany(links)
		result, err := storage.GetAll()

		if err != nil {
			t.Fatalf("GetAll() error = %v, want nil", err)
		}
		if len(result) != 1 {
			t.Fatalf("GetAll() returned %d groups, want 1", len(result))
		}

		retrievedLinks := result[0].Links
		if len(retrievedLinks) != len(links) {
			t.Fatalf("GetAll() returned %d links, want %d", len(retrievedLinks), len(links))
		}

		if retrievedLinks[0].URL != links[0].URL {
			t.Errorf("GetAll() first link URL = %s, want %s", retrievedLinks[0].URL, links[0].URL)
		}
		if retrievedLinks[0].Status != links[0].Status {
			t.Errorf("GetAll() first link Status = %s, want %s", retrievedLinks[0].Status, links[0].Status)
		}
	})

	t.Run("concurrent reads return correct data", func(t *testing.T) {
		storage := New()
		const numGroups = 5

		// Создаем несколько групп
		for i := 0; i < numGroups; i++ {
			links := []models.Link{
				createTestLink("https://example.com", models.LinkStatusAvailable),
			}
			_, _ = storage.InsertMany(links)
		}

		// Параллельно читаем все группы
		const numReaders = 10
		var wg sync.WaitGroup
		errs := make(chan error, numReaders)

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := storage.GetAll()
				if err != nil {
					errs <- err
					return
				}
				if len(result) != numGroups {
					errs <- err
					return
				}
			}()
		}

		wg.Wait()
		close(errs)

		if len(errs) > 0 {
			t.Fatalf("GetAll() returned %d errors during concurrent reads", len(errs))
		}
	})
}
