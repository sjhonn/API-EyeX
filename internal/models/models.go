package models

type Palette struct {
	Background string `json:"background"`
	Surface    string `json:"surface"`
	Text       string `json:"text"`
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Error      string `json:"error"`
	Success    string `json:"success"`
}

type ThemeResponse struct {
	Type    string  `json:"type"`
	Palette Palette `json:"palette"`
}

type TypesResponse struct {
	Types []string `json:"types"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
