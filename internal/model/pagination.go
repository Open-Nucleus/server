package model

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

const (
	DefaultPage    = 1
	DefaultPerPage = 25
	MaxPerPage     = 100
)

// PaginationFromRequest extracts pagination parameters from query strings.
func PaginationFromRequest(r *http.Request) (page, perPage int) {
	page = queryInt(r, "page", DefaultPage)
	perPage = queryInt(r, "per_page", DefaultPerPage)

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return page, perPage
}

// NewPagination creates a Pagination with computed total pages.
func NewPagination(page, perPage, total int) *Pagination {
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	return &Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
