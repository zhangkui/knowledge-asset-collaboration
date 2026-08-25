package pagination

import (
	"net/url"
	"strconv"
)

type Request struct {
	Page     int
	PageSize int
}
type Response struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Items    any `json:"items"`
}

func FromQuery(q url.Values) Request {
	page := number(q.Get("page"), 1)
	size := number(q.Get("page_size"), 20)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return Request{Page: page, PageSize: size}
}
func (r Request) Offset() int { return (r.Page - 1) * r.PageSize }
func (r Request) Limit(total int) (int, int) {
	start := r.Offset()
	if start > total {
		start = total
	}
	end := start + r.PageSize
	if end > total {
		end = total
	}
	return start, end
}
func number(v string, f int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return f
	}
	return n
}
