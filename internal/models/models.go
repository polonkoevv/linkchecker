package models

import "time"

type LinkStatus string

const (
	LinkStatusAvailable    LinkStatus = "available"
	LinkStatusNotAvailable LinkStatus = "not available"
)

type Links struct {
	Links    []Link `json:"links"`
	LinksNum int    `json:"links_num"`
}

type Link struct {
	URL       string        `json:"url"`
	Status    LinkStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	CheckedAt time.Time     `json:"checked_at"`
}

type LinksResponse struct {
	Links    map[string]LinkStatus `json:"links"`
	LinksNum int                   `json:"links_num"`
}

type GenerateReportRequest struct {
	LinksNum []int `json:"links_num"`
}

type GenerateReportResponse struct {
	Message string `json:"message"`
	Size    int    `json:"size_bytes"`
}
