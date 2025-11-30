package link

import (
	"bytes"
	"context"
	"log/slog"
	"sync"

	"github.com/polonkoevv/linkchecker/internal/models"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
	"github.com/polonkoevv/linkchecker/internal/urlchecker"
)

type linkRepository interface {
	InsertMany(links []models.Link) (int, error)
	GetByNums(linksNum []int) ([]models.Links, error)
	GetAll() ([]models.Links, error)
}

// LinkService contains business logic for checking links and generating reports.
type Service struct {
	repository   linkRepository
	urlChecker   *urlchecker.Checker
	pdfGenerator *pdfgenerator.GoFPDFGenerator

	workerCount int
}

const defaultWorkerCount = 4

// New creates a LinkService with the given repository, PDF generator and worker pool size.
func New(repo linkRepository, workerCount int) *Service {
	if workerCount <= 0 {
		workerCount = defaultWorkerCount
	}

	return &Service{
		repository:   repo,
		urlChecker:   urlchecker.NewChecker(),
		pdfGenerator: pdfgenerator.NewGoFPDFGenerator(),
		workerCount:  workerCount,
	}
}

// deduplicateLinks removes duplicate links from the slice.
func deduplicateLinks(links []string) []string {
	seen := make(map[string]struct{}, len(links))
	unique := make([]string, 0, len(links))

	for _, raw := range links {
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		unique = append(unique, raw)
	}

	return unique
}

// startWorkers launches worker goroutines to check URLs.
func (s *Service) startWorkers(ctx context.Context, jobs <-chan string, results chan<- models.Link, workerCount int) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func(id int) {
			defer wg.Done()
			s.worker(ctx, id, jobs, results)
		}(i)
	}

	return &wg
}

// worker processes URLs from jobs channel and sends results.
func (s *Service) worker(ctx context.Context, id int, jobs <-chan string, results chan<- models.Link) {
	for raw := range jobs {
		if ctx.Err() != nil {
			slog.Warn("worker exiting due to context done", slog.Int("worker_id", id))
			return
		}

		link := s.urlChecker.CheckURLWithContext(ctx, raw)

		select {
		case <-ctx.Done():
			slog.Warn("worker canceled while sending result", slog.Int("worker_id", id))
			return
		case results <- link:
		}
	}
}

// startProducer sends links to jobs channel.
func (s *Service) startProducer(ctx context.Context, jobs chan<- string, links []string) {
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
}

// buildResponse creates LinksResponse from checked links.
func (s *Service) buildResponse(checkedLinks []models.Link, linksNum int) models.LinksResponse {
	res := models.LinksResponse{
		Links:    make(map[string]models.LinkStatus, len(checkedLinks)),
		LinksNum: linksNum,
	}
	for _, l := range checkedLinks {
		res.Links[l.URL] = l.Status
	}
	return res
}

// collectResults collects results from channel until it's closed.
func (s *Service) collectResults(ctx context.Context, results <-chan models.Link) ([]models.Link, error) {
	checkedLinks := make([]models.Link, 0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case link, ok := <-results:
			if !ok {
				return checkedLinks, nil
			}
			checkedLinks = append(checkedLinks, link)
		}
	}
}

// CheckMany validates and checks the given links concurrently using a worker pool.
func (s *Service) CheckMany(ctx context.Context, links []string) (models.LinksResponse, error) {
	unique := deduplicateLinks(links)
	linksLen := len(unique)

	if linksLen == 0 {
		return models.LinksResponse{
			Links:    map[string]models.LinkStatus{},
			LinksNum: 0,
		}, nil
	}

	slog.Info("checking links with worker pool", slog.Int("count", linksLen))

	workerCount := s.workerCount
	if workerCount > linksLen {
		workerCount = linksLen
	}

	jobs := make(chan string)
	results := make(chan models.Link)

	wg := s.startWorkers(ctx, jobs, results, workerCount)
	s.startProducer(ctx, jobs, unique)

	go func() {
		wg.Wait()
		close(results)
	}()

	checkedLinks, err := s.collectResults(ctx, results)
	if err != nil {
		slog.Warn("check many canceled by context")
		return models.LinksResponse{}, err
	}

	linksNum, err := s.repository.InsertMany(checkedLinks)
	if err != nil {
		slog.Error("failed to insert checked links", slog.Any("error", err))
		return models.LinksResponse{}, err
	}

	res := s.buildResponse(checkedLinks, linksNum)

	slog.Debug("links checked and stored with worker pool",
		slog.Int("links_num", linksNum),
		slog.Int("links_count", len(checkedLinks)),
		slog.Int("workers", workerCount),
	)

	return res, nil
}

// GenerateReport builds a PDF report for the specified link group numbers.
func (s *Service) GenerateReport(ctx context.Context, linksNum []int) (*bytes.Buffer, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	slog.Info("generating report for links groups", slog.Int("groups", len(linksNum)))

	checkedLinks, err := s.repository.GetByNums(linksNum)
	if err != nil {
		slog.Error("failed to get links by nums", slog.Any("error", err))
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
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

// GetAll returns all stored link groups from the repository.
func (s *Service) GetAll(ctx context.Context) ([]models.Links, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	slog.Info("fetching all links groups")

	allLinks, err := s.repository.GetAll()
	if err != nil {
		slog.Error("failed to get all links", slog.Any("error", err))
		return nil, err
	}

	slog.Debug("fetched all links groups", slog.Int("groups_count", len(allLinks)))

	return allLinks, nil
}
