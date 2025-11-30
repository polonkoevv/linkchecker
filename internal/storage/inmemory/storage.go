package inmemory

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/polonkoevv/linkchecker/internal/models"
)

// Storage implements an in-memory link repository with optional JSON persistence.
type Storage struct {
	links map[int][]models.Link
	mtx   sync.RWMutex
}

// New creates an empty in-memory Storage instance.
func New() *Storage {
	return &Storage{
		links: make(map[int][]models.Link),
		mtx:   sync.RWMutex{},
	}
}

// InsertMany stores a batch of links and returns its group number.
func (s *Storage) InsertMany(links []models.Link) (int, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	num := len(s.links) + 1
	s.links[num] = links

	slog.Debug("inserted links batch",
		slog.Int("links_num", num),
		slog.Int("links_count", len(links)),
	)

	return num, nil
}

// GetByNums returns stored link groups for the given group numbers.
func (s *Storage) GetByNums(linksNum []int) ([]models.Links, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	res := []models.Links{}

	for _, num := range linksNum {
		links, ok := s.links[num]
		if !ok {
			slog.Warn("requested links_num not found", slog.Int("links_num", num))
			return nil, errors.New("invalid link number")
		}
		res = append(res, models.Links{
			LinksNum: num,
			Links:    links,
		})
	}

	slog.Debug("loaded links by nums", slog.Int("requested_groups", len(linksNum)), slog.Int("returned_groups", len(res)))

	return res, nil
}

// GetAll returns all stored link groups.
func (s *Storage) GetAll() ([]models.Links, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	res := []models.Links{}

	for k, v := range s.links {
		res = append(res, models.Links{
			LinksNum: k,
			Links:    v,
		})
	}

	slog.Debug("loaded all links groups", slog.Int("groups_count", len(res)))

	return res, nil
}

// LoadFromFile populates storage state from a JSON file if it exists.
func (s *Storage) LoadFromFile(path string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing file is not an error, there is simply nothing to load yet
			return nil
		}
		return fmt.Errorf("open storage file: %w", err)
	}
	defer file.Close()

	var groups []models.Links
	if err := json.NewDecoder(file).Decode(&groups); err != nil {
		return fmt.Errorf("decode storage file: %w", err)
	}

	s.links = make(map[int][]models.Link, len(groups))
	for _, g := range groups {
		s.links[g.LinksNum] = g.Links
	}

	return nil
}

// SaveToFile writes current storage state to a JSON file atomically.
func (s *Storage) SaveToFile(path string) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	groups := make([]models.Links, 0, len(s.links))
	for num, links := range s.links {
		groups = append(groups, models.Links{
			LinksNum: num,
			Links:    links,
		})
	}

	tmpPath := path + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create storage file: %w", err)
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	if err := enc.Encode(groups); err != nil {
		file.Close()
		return fmt.Errorf("encode storage file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close storage file: %w", err)
	}

	// атомарная замена: tmp → основной файл
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename storage file: %w", err)
	}

	return nil
}
