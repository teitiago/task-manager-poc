package models

// Pagination allow to limit the number of results returned over a result.
type Pagination struct {
	Limit int
	Page  int
	Sort  string
}
