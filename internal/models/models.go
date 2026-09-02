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
	Type       string  `json:"type"`
	Palette    Palette `json:"palette"`
	ContrastOK bool    `json:"contrast_ok"`
}

type TypesResponse struct {
	Types []string `json:"types"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type CustomThemeRequest struct {
	Type         string  `json:"type"`
	Palette      Palette `json:"palette"`
	Severity     string  `json:"severity,omitempty"`
	Mode         string  `json:"mode,omitempty"`
	HighContrast bool    `json:"high_contrast,omitempty"`
}

type QuickTestAnswers struct {
	RedsLookDarker      bool `json:"reds_look_darker"`
	GreenBrownConfusion bool `json:"green_brown_confusion"`
	BlueYellowConfusion bool `json:"blue_yellow_confusion"`
	ColorsLookGray      bool `json:"colors_look_gray"`
}

type QuickTestRequest struct {
	Answers QuickTestAnswers `json:"answers"`
}

type QuickTestResponse struct {
	SuggestedType string `json:"suggested_type"`
	Disclaimer    string `json:"disclaimer"`
}

type SimulateRequest struct {
	Hex      string   `json:"hex"`
	Type     string   `json:"type"`
	Severity *float64 `json:"severity,omitempty"`
}

type SimulateResponse struct {
	Original  string  `json:"original"`
	Simulated string  `json:"simulated"`
	Type      string  `json:"type"`
	Severity  float64 `json:"severity"`
	Model     string  `json:"model"`
}

type SimulatedColor struct {
	Original  string `json:"original"`
	Simulated string `json:"simulated"`
}

type SimulateBatchRequest struct {
	Colors   []string `json:"colors"`
	Type     string   `json:"type"`
	Severity *float64 `json:"severity,omitempty"`
}

type SimulateBatchResponse struct {
	Type     string           `json:"type"`
	Severity float64          `json:"severity"`
	Model    string           `json:"model"`
	Results  []SimulatedColor `json:"results"`
}
