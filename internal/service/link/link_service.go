package link

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
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

	workerCount int
}

const defaultWorkerCount = 4

func New(repo linkRepository, timeout time.Duration, pdfGenerator *pdfgenerator.GoFPDFGenerator, workerCount int) *LinkService {
	if workerCount <= 0 {
		workerCount = defaultWorkerCount
	}

	return &LinkService{
		repository:   repo,
		urlChecker:   urlchecker.NewChecker(timeout),
		pdfGenerator: pdfGenerator,
		workerCount:  workerCount,
	}
}

func (s *LinkService) CheckMany(ctx context.Context, links []string) (models.LinksResponse, error) {
	linksLen := len(links)
	checkedLinks := make([]models.Link, 0, linksLen)

	slog.Info("checking links with worker pool", slog.Int("count", linksLen))

	if linksLen == 0 {
		return models.LinksResponse{
			Links:    map[string]models.LinkStatus{},
			LinksNum: 0,
		}, nil
	}

	jobs := make(chan string)
	results := make(chan models.Link)

	workerCount := s.workerCount
	if workerCount > linksLen {
		workerCount = linksLen
	}

	var wgWorkers sync.WaitGroup
	wgWorkers.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func(id int) {
			defer wgWorkers.Done()

			for raw := range jobs {
				select {
				case <-ctx.Done():
					slog.Warn("worker exiting due to context done", slog.Int("worker_id", id))
					return
				default:
				}

				link := s.urlChecker.CheckURLWithContext(ctx, raw)

				select {
				case <-ctx.Done():
					slog.Warn("worker canceled while sending result", slog.Int("worker_id", id))
					return
				case results <- link:
				}
			}
		}(i)
	}

	go func() {
		defer close(jobs)
		for _, raw := range links {
			select {
			case <-ctx.Done():
				slog.Warn("producer stopped due to context done")
				return
			case jobs <- raw:
			}
		}
	}()

	go func() {
		wgWorkers.Wait()
		close(results)
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Warn("check many canceled by context")
			return models.LinksResponse{}, ctx.Err()

		case link, ok := <-results:
			if !ok {
				linksNum, err := s.repository.InsertMany(checkedLinks)
				if err != nil {
					slog.Error("failed to insert checked links", slog.Any("error", err))
					return models.LinksResponse{}, err
				}

				res := models.LinksResponse{
					Links:    make(map[string]models.LinkStatus, len(checkedLinks)),
					LinksNum: linksNum,
				}
				for _, l := range checkedLinks {
					res.Links[l.URL] = l.Status
				}

				slog.Debug("links checked and stored with worker pool",
					slog.Int("links_num", linksNum),
					slog.Int("links_count", len(checkedLinks)),
					slog.Int("workers", workerCount),
				)

				return res, nil
			}

			checkedLinks = append(checkedLinks, link)
		}
	}
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
