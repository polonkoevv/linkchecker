package link

import (
	"bytes"
	"context"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
	"github.com/polonkoevv/linkchecker/internal/urlchecker"
)

type linkRepository interface {
	InsertMany(links []models.Link) (int, error)
	GetByNums(links_num []int) ([]models.Links, error)
	GetAll() ([]models.Links, error)
}

type LinkService struct {
	repository   linkRepository
	urlChecker   *urlchecker.Checker
	pdfGenerator *pdfgenerator.GoFPDFGenerator
}

func New(repo linkRepository, timeout time.Duration, pdfGenerator *pdfgenerator.GoFPDFGenerator) *LinkService {
	return &LinkService{
		repository:   repo,
		urlChecker:   urlchecker.NewChecker(timeout),
		pdfGenerator: pdfGenerator,
	}
}

func (s *LinkService) CheckMany(ctx context.Context, links []string) (models.LinksResponse, error) {

	linksLen := len(links)

	checkedLinks := make([]models.Link, 0, linksLen)

	for _, raw := range links {
		select {
		case <-ctx.Done():
			return models.LinksResponse{}, ctx.Err()
		default:
		}

		// checkedLinks = append(checkedLinks, s.urlChecker.CheckURLWithContext(ctx, raw))
		checkedLinks = append(checkedLinks, s.urlChecker.CheckURL(raw))
	}

	linksNum, err := s.repository.InsertMany(checkedLinks)
	if err != nil {
		return models.LinksResponse{}, err
	}

	res := models.LinksResponse{
		Links: make(map[string]models.LinkStatus, len(checkedLinks)),
	}

	for _, link := range checkedLinks {
		res.Links[link.URL] = link.Status
	}

	res.LinksNum = linksNum

	return res, nil
}

func (s *LinkService) GenerateReport(ctx context.Context, links_num []int) (*bytes.Buffer, error) {

	checkedLinks, err := s.repository.GetByNums(links_num)
	if err != nil {
		return nil, err
	}

	report, err := s.pdfGenerator.GenerateMultipleReports(checkedLinks)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s *LinkService) GetAll(ctx context.Context) ([]models.Links, error) {
	allLinks, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}

	return allLinks, nil
}
