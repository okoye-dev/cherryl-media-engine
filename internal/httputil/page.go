package httputil

import (
	"net/http"
	"strconv"
)

// Page is one page of a list: which page, how many per page, and where it starts.
type Page struct {
	Number int
	Size   int
	Offset int
}

// ParsePage reads ?page= and ?pageSize= (the older ?limit= still works as the page size).
// Pages start at 1. The size is never above max, so no request can ask for everything at once.
func ParsePage(r *http.Request, defaultSize, maxSize int) Page {
	q := r.URL.Query()
	size := defaultSize
	for _, key := range []string{"pageSize", "limit"} {
		if v, err := strconv.Atoi(q.Get(key)); err == nil && v > 0 {
			size = v
			break
		}
	}
	if size > maxSize {
		size = maxSize
	}
	number := 1
	if v, err := strconv.Atoi(q.Get("page")); err == nil && v > 0 {
		number = v
	}
	return Page{Number: number, Size: size, Offset: (number - 1) * size}
}

// TotalPages is how many pages a list of total items has.
func (p Page) TotalPages(total int) int {
	if total <= 0 || p.Size <= 0 {
		return 0
	}
	return (total + p.Size - 1) / p.Size
}

// SetPageHeaders tells the client where it is in the list. The body stays a plain array so older clients keep working.
func SetPageHeaders(w http.ResponseWriter, total int, p Page) {
	h := w.Header()
	h.Set("X-Total-Count", strconv.Itoa(total))
	h.Set("X-Page", strconv.Itoa(p.Number))
	h.Set("X-Page-Size", strconv.Itoa(p.Size))
	h.Set("X-Total-Pages", strconv.Itoa(p.TotalPages(total)))
}
