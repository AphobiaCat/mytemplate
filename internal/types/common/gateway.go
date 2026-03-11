package common

type LanguageItem struct {
	Lang     string `json:"lang"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type ListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"` // desc or asc
	Type     string `json:"type"`     // card or banner
}
