package api

type Link struct {
	Href string `json:"href"`
}

type ActionField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Action struct {
	Href        string        `json:"href"`
	Method      string        `json:"method"`
	ContentType string        `json:"content_type"`
	Fields      []ActionField `json:"fields"`
}
