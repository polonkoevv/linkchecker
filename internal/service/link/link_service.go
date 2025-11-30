package link

import (
	"bytes"
	"context"
	"log/slog"
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

	slog.Info("checking links", slog.Int("count", linksLen))

	for _, raw := range links {
		select {
		case <-ctx.Done():
			slog.Warn("check many canceled by context")
			return models.LinksResponse{}, ctx.Err()
		default:
		}

		// checkedLinks = append(checkedLinks, s.urlChecker.CheckURLWithContext(ctx, raw))
		checkedLinks = append(checkedLinks, s.urlChecker.CheckURL(raw))
	}

	linksNum, err := s.repository.InsertMany(checkedLinks)
	if err != nil {
		slog.Error("failed to insert checked links", slog.Any("error", err))
		return models.LinksResponse{}, err
	}

	res := models.LinksResponse{
		Links: make(map[string]models.LinkStatus, len(checkedLinks)),
	}

	for _, link := range checkedLinks {
		res.Links[link.URL] = link.Status
	}

	res.LinksNum = linksNum

	slog.Debug("links checked and stored",
		slog.Int("links_num", linksNum),
		slog.Int("links_count", len(checkedLinks)),
	)

	return res, nil
}

func (s *LinkService) GenerateReport(ctx context.Context, linksNum []int) (*bytes.Buffer, error) {
	slog.Info("generating report for links groups", slog.Int("groups", len(linksNum)))

	checkedLinks, err := s.repository.GetByNums(linksNum)
	if err != nil {
		slog.Error("failed to get links by nums", slog.Any("error", err))
		return nil, err
	}

	report, err := s.pdfGenerator.GenerateMultipleReports(checkedLinks)
	if err != nil {
		slog.Error("failed to generate PDF report", slog.Any("error", err))
		return nil, err
	}

	slog.Debug("PDF report generated successfully",
		slog.Int("groups", len(linksNum)),
	)

	return report, nil
}

func (s *LinkService) GetAll(ctx context.Context) ([]models.Links, error) {
	slog.Info("fetching all links groups")

	allLinks, err := s.repository.GetAll()
	if err != nil {
		slog.Error("failed to get all links", slog.Any("error", err))
		return nil, err
	}

	slog.Debug("fetched all links groups", slog.Int("groups_count", len(allLinks)))

	return allLinks, nil
}
