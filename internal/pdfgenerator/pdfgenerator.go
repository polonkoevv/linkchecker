package pdfgenerator

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/polonkoevv/linkchecker/internal/models"
)

// GoFPDFGenerator generates PDF reports using gofpdf
type GoFPDFGenerator struct {
}

type pdfStatistic struct {
	available                int
	notAvailable             int
	averageAvailableSpeed    time.Duration
	averageNotAvailableSpeed time.Duration
	total                    int
}

const title = "LINK STATUS REPORT - GROUP"

// Page settings
const orientationStr string = "P"
const unitStr string = "mm"
const sizeStr string = "A4"
const fontDirStr string = ""

// Font
const familyStr string = "Arial"
const styleStr string = "B"
const size float64 = 20

// NewGoFPDFGenerator creates a new GoFPDFGenerator instance.
func NewGoFPDFGenerator() *GoFPDFGenerator {
	return &GoFPDFGenerator{}
}

// GenerateReport builds a single-group PDF report for the given links.
func (g *GoFPDFGenerator) GenerateReport(links models.Links) (*bytes.Buffer, error) {
	slog.Info("generating single PDF report",
		slog.Int("links_num", links.LinksNum),
		slog.Int("links_count", len(links.Links)),
	)

	pdf := gofpdf.New(orientationStr, unitStr, sizeStr, fontDirStr)
	pdf.AddPage()

	// Добавляем заголовок
	g.addHeaderWithGroup(pdf, links.LinksNum)

	// Рассчитываем статистику
	stats := g.calculateStatistic(links)

	// Добавляем статистику в отчет
	g.addStatistics(pdf, stats)

	// Добавляем детальную информацию по ссылкам
	g.addDetailedLinks(pdf, links)

	// Создаем буфер в памяти
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		slog.Error("failed to generate single PDF report", slog.Any("error", err))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	slog.Debug("single PDF report generated",
		slog.Int("links_num", links.LinksNum),
		slog.Int("size_bytes", buf.Len()),
	)

	return &buf, nil
}

// GenerateMultipleReports builds a multi-page PDF for several link groups.
func (g *GoFPDFGenerator) GenerateMultipleReports(linksSlice []models.Links) (*bytes.Buffer, error) {
	slog.Info("generating multi-group PDF report", slog.Int("groups", len(linksSlice)))

	pdf := gofpdf.New(orientationStr, unitStr, sizeStr, fontDirStr)

	for _, links := range linksSlice {
		pdf.AddPage()

		g.addHeaderWithGroup(pdf, links.LinksNum)

		stats := g.calculateStatistic(links)

		g.addStatistics(pdf, stats)

		g.addDetailedLinks(pdf, links)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		slog.Error("failed to generate multi-group PDF report", slog.Any("error", err))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	slog.Debug("multi-group PDF report generated",
		slog.Int("groups", len(linksSlice)),
		slog.Int("size_bytes", buf.Len()),
	)

	return &buf, nil
}

func (g *GoFPDFGenerator) addHeaderWithGroup(pdf *gofpdf.Fpdf, groupNum int) {
	pdf.SetFont(familyStr, styleStr, size)
	pdf.SetTextColor(0, 0, 128)
	pdf.CellFormat(0, 15, fmt.Sprintf("%s %d", title, groupNum), "", 0, "C", false, 0, "")
	pdf.Ln(20)
}

func (g *GoFPDFGenerator) calculateStatistic(links models.Links) *pdfStatistic {
	res := &pdfStatistic{}
	res.total = len(links.Links)

	for _, link := range links.Links {
		if link.Status == models.LinkStatusAvailable {
			res.available++
			res.averageAvailableSpeed += link.Duration
		} else {
			res.notAvailable++
			res.averageNotAvailableSpeed += link.Duration
		}
	}

	if res.available > 0 {
		res.averageAvailableSpeed = res.averageAvailableSpeed / time.Duration(res.available)
	}
	if res.notAvailable > 0 {
		res.averageNotAvailableSpeed = res.averageNotAvailableSpeed / time.Duration(res.notAvailable)
	}

	return res
}

func (g *GoFPDFGenerator) addStatistics(pdf *gofpdf.Fpdf, stats *pdfStatistic) {
	pdf.SetFont(familyStr, styleStr, 16)
	pdf.CellFormat(0, 10, "STATISTICS SUMMARY", "", 0, "L", false, 0, "")
	pdf.Ln(12)

	pdf.SetFont(familyStr, styleStr, 12)
	pdf.SetFillColor(240, 240, 240)

	pdf.CellFormat(80, 8, "Metric", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Count", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, "Average Time", "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	pdf.SetFont(familyStr, "", 12)
	pdf.SetFillColor(255, 255, 255)

	pdf.CellFormat(80, 8, "Available Links", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%d", stats.available), "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, stats.averageAvailableSpeed.Round(time.Millisecond).String(), "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	pdf.CellFormat(80, 8, "Not Available Links", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%d", stats.notAvailable), "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, stats.averageNotAvailableSpeed.Round(time.Millisecond).String(), "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	pdf.SetFont(familyStr, styleStr, 12)
	pdf.CellFormat(80, 8, "TOTAL", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%d", stats.total), "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, "-", "1", 0, "C", true, 0, "")
	pdf.Ln(20)
}

func (g *GoFPDFGenerator) addDetailedLinks(pdf *gofpdf.Fpdf, links models.Links) {
	pdf.SetFont(familyStr, styleStr, 16)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 10, "DETAILED LINK REPORT", "", 0, "L", false, 0, "")
	pdf.Ln(12)

	pdf.SetFont(familyStr, styleStr, 10)
	pdf.SetFillColor(200, 200, 200)

	widths := []float64{60, 25, 25, 30, 40}

	pdf.CellFormat(widths[0], 8, "URL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(widths[1], 8, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(widths[2], 8, "Duration", "1", 0, "C", true, 0, "")
	pdf.CellFormat(widths[3], 8, "Checked At", "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	pdf.SetFont(familyStr, "", 8)
	fill := false

	for _, link := range links.Links {
		if fill {
			pdf.SetFillColor(240, 240, 240)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.CellFormat(widths[0], 6, truncateString(link.URL, 50), "1", 0, "L", fill, 0, "")

		statusColor := getStatusColor(link.Status)
		pdf.SetTextColor(statusColor[0], statusColor[1], statusColor[2])
		pdf.CellFormat(widths[1], 6, string(link.Status), "1", 0, "C", fill, 0, "")
		pdf.SetTextColor(0, 0, 0)

		pdf.CellFormat(widths[2], 6, link.Duration.Round(time.Millisecond).String(), "1", 0, "C", fill, 0, "")

		checkedTime := link.CheckedAt.Format("15:04:05 02.01.2006")
		pdf.CellFormat(widths[3], 6, checkedTime, "1", 0, "C", fill, 0, "")

		pdf.Ln(6)
		fill = !fill

		if pdf.GetY() > 260 {
			pdf.AddPage()
			pdf.SetFont(familyStr, styleStr, 10)
			pdf.SetFillColor(200, 200, 200)
			pdf.CellFormat(widths[0], 8, "URL", "1", 0, "C", true, 0, "")
			pdf.CellFormat(widths[1], 8, "Status", "1", 0, "C", true, 0, "")
			pdf.CellFormat(widths[2], 8, "Duration", "1", 0, "C", true, 0, "")
			pdf.CellFormat(widths[3], 8, "Checked At", "1", 0, "C", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont(familyStr, "", 8)
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getStatusColor(status models.LinkStatus) [3]int {
	switch status {
	case models.LinkStatusAvailable:
		return [3]int{0, 128, 0} // Green
	case models.LinkStatusNotAvailable:
		return [3]int{255, 0, 0} // Red
	default:
		return [3]int{0, 0, 0} // Black
	}
}
