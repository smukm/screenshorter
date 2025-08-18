package service

type Screenshot interface {
	Make(html string, opts ScreenshotOptions) ([]byte, string, error)
}

type SelectionArea struct {
	X       int `json:"x"`                 // Координата X начальной точки
	Y       int `json:"y"`                 // Координата Y начальной точки
	Width   int `json:"width"`             // Ширина выделенной области
	Height  int `json:"height"`            // Высота выделенной области
	ScrollX int `json:"scrollx,omitempty"` // Горизонтальная прокрутка
	ScrollY int `json:"scrolly,omitempty"` // Вертикальная прокрутка
}

// SelectionStyle стиль выделения
type SelectionStyle struct {
	BorderColor string  `json:"borderColor"` // Цвет рамки (CSS-формат)
	BorderWidth int     `json:"borderWidth"` // Толщина рамки (px)
	BorderStyle string  `json:"borderStyle"` // Стиль рамки: "solid", "dashed", "dotted"
	Opacity     float64 `json:"opacity"`     // Прозрачность (0.0 - 1.0)
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
	Timeout        float64 `json:"timeout"`
	Selections     []SelectionArea
	SelectionStyle *SelectionStyle
}

type Service struct {
	Screenshot Screenshot
}

func NewService(s Screenshot) *Service {
	return &Service{
		Screenshot: s,
	}
}
