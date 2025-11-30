package link

import "testing"

func TestService_New(t *testing.T) {
	t.Run("creates service with valid worker count", func(t *testing.T) {
		repo := &mockRepository{}
		service := New(repo, 5)

		if service == nil {
			t.Fatal("New() returned nil")
		}
		if service.workerCount != 5 {
			t.Errorf("New() workerCount = %d, want 5", service.workerCount)
		}
		if service.repository != repo {
			t.Error("New() repository not set correctly")
		}
	})

	t.Run("uses default worker count for zero or negative", func(t *testing.T) {
		repo := &mockRepository{}

		service1 := New(repo, 0)
		if service1.workerCount != defaultWorkerCount {
			t.Errorf("New(0) workerCount = %d, want %d", service1.workerCount, defaultWorkerCount)
		}

		service2 := New(repo, -1)
		if service2.workerCount != defaultWorkerCount {
			t.Errorf("New(-1) workerCount = %d, want %d", service2.workerCount, defaultWorkerCount)
		}
	})
}
