package inmemory

import (
	"errors"
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

	return num, nil
}

func (s *Storage) GetByNums(links_num []int) ([]models.Links, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	res := []models.Links{}

	for _, num := range links_num {
		links, ok := s.links[num]
		if !ok {
			return nil, errors.New("invalid link number")
		}
		res = append(res, models.Links{
			LinksNum: num,
			Links:    links,
		})
	}

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

	return res, nil
}
