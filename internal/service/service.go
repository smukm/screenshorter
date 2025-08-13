package service

type Screenshot interface {
	Make(html string, opts ScreenshotOptions) ([]byte, string, error)
}

// ScreenshotOptions параметры для настройки скриншота
type ScreenshotOptions struct {
	Browser        BrowserType `json:"browser"` // Тип браузера
	Quality        *int        `json:"quality"`
	Type           string      `json:"type"`
	FullPage       bool        `json:"full_page"`
	OmitBackground bool        `json:"omit_background"`
	Viewport       *struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"viewport"`
	Timeout float64 `json:"timeout"`
}

type Service struct {
	Screenshot Screenshot
}

func NewService(s Screenshot) *Service {
	return &Service{
		Screenshot: s,
	}
}
