package clipboard

import (
	"fmt"
	"io"
	"strings"
)

type History struct {
	records []Record // Oldest to latest.
}

type Record struct {
	value      string
	shortValue string

	private bool
}

func NewHistory() *History {
	return &History{}
}

func (h *History) Value(index int) string {
	record := h.records[len(h.records)-index-1]
	if record.private {
		return "[private]"
	}
	return record.value
}

func (h *History) WriteShortValues(w io.Writer) {
	for i := len(h.records) - 1; i >= 0; i-- {
		fmt.Fprintf(w, "%v: %s\n", len(h.records)-i, h.records[i].shortValue)
	}
}

func (h *History) LatestValue() string {
	if len(h.records) == 0 {
		return ""
	}
	return h.records[len(h.records)-1].value
}

func (h *History) Len() int {
	return len(h.records)
}

func (h *History) Save(value string) bool {
	if value == h.LatestValue() {
		return false
	}

	// Move up any duplicates.
	for i, record := range h.records {
		if value == record.value {
			h.records = append(append(h.records[:i], h.records[i+1:]...), record)
			return true
		}
	}

	shortValue := ""
	if len(value) > 45 {
		shortValue = value[0:45]
	} else {
		shortValue = value
	}

	shortValue = strings.Replace(shortValue, "\n", "↵", -1)
	shortValue = strings.Replace(shortValue, "  ", " ", -1)

	h.records = append(h.records, Record{
		value:      value,
		shortValue: shortValue,
	})

	return true
}
