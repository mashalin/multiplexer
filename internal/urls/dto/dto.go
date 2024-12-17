package dto

type RequestBody struct {
	URLs []string `json:"urls"`
}

type ResponseData struct {
	URL  string `json:"url"`
	Body string `json:"body"`
}
