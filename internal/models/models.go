package models

import "time"

// LinkStatus describes availability status of a checked link.
type LinkStatus string

const (
	LinkStatusAvailable    LinkStatus = "available"
	LinkStatusNotAvailable LinkStatus = "not available"
)

// Links groups a slice of links with its assigned group number.
type Links struct {
	Links    []Link `json:"links"`
	LinksNum int    `json:"links_num"`
}

// Link holds the result of a single URL availability check.
type Link struct {
	URL       string        `json:"url"`
	Status    LinkStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	CheckedAt time.Time     `json:"checked_at"`
}

// LinksResponse is returned from POST /links with statuses and group id.
type LinksResponse struct {
	Links    map[string]LinkStatus `json:"links"`
	LinksNum int                   `json:"links_num"`
}

// GenerateReportRequest represents a list of link group numbers to report on.
type GenerateReportRequest struct {
	LinksNum []int `json:"links_num"`
}

// GenerateReportResponse is a JSON metadata response for generated PDF report.
type GenerateReportResponse struct {
	Message string `json:"message"`
	Size    int    `json:"size_bytes"`
}
