package urlchecker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/polonkoevv/linkchecker/internal/models"
)

// Checker performs HTTP HEAD requests to determine link availability.
type Checker struct {
	client *http.Client
}

// NewChecker creates a new Checker with a default HTTP client.
func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{},
	}
}

// CheckURL checks the given URL without external context control.
func (c *Checker) CheckURL(rawURL string) models.Link {
	start := time.Now()

	// Нормализуем URL
	normalizedURL, err := c.normalizeURL(rawURL)
	if err != nil {
		slog.Warn("failed to normalize URL",
			slog.String("raw_url", rawURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}

	// Создаем запрос с правильными заголовками
	req, err := http.NewRequest("HEAD", normalizedURL, nil)
	if err != nil {
		slog.Error("failed to create HTTP request",
			slog.String("url", normalizedURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}

	req.Header.Set("User-Agent", "WebStatusChecker/1.0")
	req.Header.Set("Accept", "*/*")

	// Выполняем запрос
	resp, err := c.client.Do(req)
	if err != nil {
		slog.Debug("HTTP request failed",
			slog.String("url", normalizedURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	// Считаем доступным если статус 2xx или 3xx
	status := models.LinkStatusNotAvailable
	if resp.StatusCode < 400 {
		status = models.LinkStatusAvailable
	}

	slog.Debug("checked URL",
		slog.String("url", rawURL),
		slog.Int("status_code", resp.StatusCode),
		slog.String("status", string(status)),
		slog.Duration("duration", duration),
	)

	return models.Link{
		URL:       rawURL,
		Status:    status,
		CheckedAt: start,
		Duration:  duration,
	}
}

// CheckURLWithContext проверяет ссылку с контекстом
func (c *Checker) CheckURLWithContext(ctx context.Context, rawURL string) models.Link {
	start := time.Now()

	normalizedURL, err := c.normalizeURL(rawURL)
	if err != nil {
		slog.Warn("failed to normalize URL",
			slog.String("raw_url", rawURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", normalizedURL, nil)
	if err != nil {
		slog.Error("failed to create HTTP request with context",
			slog.String("url", normalizedURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}

	req.Header.Set("User-Agent", "WebStatusChecker/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Debug("HTTP request with context failed",
			slog.String("url", normalizedURL),
			slog.Any("error", err),
		)
		return models.Link{
			URL:       rawURL,
			Status:    models.LinkStatusNotAvailable,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	status := models.LinkStatusNotAvailable
	if resp.StatusCode < 400 {
		status = models.LinkStatusAvailable
	}

	slog.Debug("checked URL with context",
		slog.String("url", rawURL),
		slog.Int("status_code", resp.StatusCode),
		slog.String("status", string(status)),
		slog.Duration("duration", duration),
	)

	return models.Link{
		URL:       rawURL,
		Status:    status,
		CheckedAt: start,
		Duration:  duration,
	}
}

func (c *Checker) normalizeURL(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if u.Host == "" {
		return "", fmt.Errorf("missing host in URL")
	}

	return u.String(), nil
}
