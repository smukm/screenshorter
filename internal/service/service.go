package service

type Screenshot interface {
	Make(html string, opts ScreenshotOptions) ([]byte, string, error)
}

type SelectionArea struct {
	X      int `json:"x"`      // Координата X начальной точки
	Y      int `json:"y"`      // Координата Y начальной точки
	Width  int `json:"width"`  // Ширина выделенной области
	Height int `json:"height"` // Высота выделенной области
}

// ScreenshotOptions параметры для настройки скриншота
type ScreenshotOptions struct {
	Browser        BrowserType `json:"browser"`
	Quality        *int        `json:"quality"`
	Type           string      `json:"type"`
	FullPage       bool        `json:"full_page"`
	OmitBackground bool        `json:"omit_background"`
	Viewport       *struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"viewport"`
	Timeout    float64 `json:"timeout"`
	Selections []SelectionArea
}

type Service struct {
	Screenshot Screenshot
}

func NewService(s Screenshot) *Service {
	return &Service{
		Screenshot: s,
	}
}
