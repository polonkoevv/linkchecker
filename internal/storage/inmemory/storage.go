package inmemory

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/polonkoevv/linkchecker/internal/models"
)

type Storage struct {
	links map[int][]models.Link
	mtx   sync.RWMutex
}

func New() *Storage {
	return &Storage{
		links: make(map[int][]models.Link),
		mtx:   sync.RWMutex{},
	}
}

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
