package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"bittimetable/internal/engine"

	"github.com/jung-kurt/gofpdf"
)

// Export generates a PDF from a list of ScheduleSlots and returns the byte content
func Export(slots []engine.ScheduleSlot, timetableName string) ([]byte, error) {
	f := gofpdf.New("P", "mm", "A4", "")
	f.AddPage()
	f.SetMargins(10, 10, 10)

	// Title
	f.SetFont("Arial", "B", 14)
	f.CellFormat(0, 10, timetableName+" - Exam Timetable", "0", 1, "C", false, 0, "")
	f.Ln(4)

	// Table header
	f.SetFont("Arial", "B", 9)
	f.SetFillColor(220, 220, 220)
	f.CellFormat(45, 8, "Exam Day - Session", "1", 0, "C", true, 0, "")
	f.CellFormat(145, 8, "Course Code - Title", "1", 1, "C", true, 0, "")

	for _, slot := range slots {
		dayLabel := fmt.Sprintf("Day %d - %s", slot.DayNumber, slot.Session)

		if len(slot.Courses) == 0 {
			f.SetFont("Arial", "B", 8)
			f.SetFillColor(245, 245, 245)
			f.CellFormat(45, 7, dayLabel, "1", 0, "C", true, 0, "")
			f.SetFont("Arial", "", 8)
			f.CellFormat(145, 7, "(No courses scheduled)", "1", 1, "L", false, 0, "")
			continue
		}

		for ci, sc := range slot.Courses {
			var cleanCodes []string
			for _, c := range sc.CourseCodes {
				// Absolute Alphanumeric Filter
				reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
				clean := reg.ReplaceAllString(c, "")
				if clean != "" {
					cleanCodes = append(cleanCodes, clean)
				}
			}
			codeStr := strings.Join(cleanCodes, " / ")
			line := codeStr + " - " + strings.ToUpper(sc.CourseName)
			lines := wrapText(line, 130)
			rowHeight := float64(len(lines)) * 5.5
			if rowHeight < 7 {
				rowHeight = 7
			}

			if ci == 0 {
				f.SetFont("Arial", "B", 8)
				f.SetFillColor(240, 248, 255)
				x, y := f.GetXY()
				f.MultiCell(45, rowHeight, dayLabel, "1", "C", true)
				f.SetXY(x+45, y)
			} else {
				f.SetFont("Arial", "", 8)
				x, y := f.GetXY()
				f.CellFormat(45, rowHeight, "", "1", 0, "", false, 0, "")
				f.SetXY(x+45, y)
			}

			f.SetFont("Arial", "", 8)
			f.SetFillColor(255, 255, 255)
			f.MultiCell(145, rowHeight/float64(len(lines)), line, "1", "L", false)
		}
	}

	var buf bytes.Buffer
	err := f.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, w := range words {
		if len(current)+len(w)+1 > maxWidth {
			if current != "" {
				lines = append(lines, current)
			}
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
