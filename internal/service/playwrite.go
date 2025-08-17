package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/playwright-community/playwright-go"
)

type BrowserType string

const (
	BrowserChromium BrowserType = "chromium"
	BrowserFirefox  BrowserType = "firefox"
	BrowserWebkit   BrowserType = "webkit"
)

type Playwrite struct {
	pw *playwright.Playwright
}

func NewPlaywrite() (*Playwrite, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not launch playwright: %w", err)
	}
	return &Playwrite{pw: pw}, nil
}

// Close освобождает ресурсы
func (p *Playwrite) Close() error {
	return p.pw.Stop()
}

func (p *Playwrite) Make(html string, opts ScreenshotOptions) ([]byte, string, error) {
	// Выбираем браузер в зависимости от параметра
	var browser playwright.Browser
	var err error

	switch opts.Browser {
	case BrowserFirefox:
		browser, err = p.pw.Firefox.Launch()
	case BrowserWebkit:
		browser, err = p.pw.WebKit.Launch()
	default: // По умолчанию Chromium
		browser, err = p.pw.Chromium.Launch()
	}

	if err != nil {
		return nil, "", fmt.Errorf("could not launch %s browser: %w", opts.Browser, err)
	}
	defer browser.Close()

	// Создаем временный HTML файл
	tmpDir := os.TempDir()
	htmlPath := filepath.Join(tmpDir, "screenshot.html")

	err = os.WriteFile(htmlPath, []byte(html), 0644)
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(htmlPath)

	page, err := browser.NewPage()
	if err != nil {
		return nil, "", err
	}
	defer page.Close()

	// Устанавливаем размер viewport если указан
	if opts.Viewport != nil {
		if err := page.SetViewportSize(opts.Viewport.Width, opts.Viewport.Height); err != nil {
			return nil, "", err
		}
	}

	// Загружаем локальный HTML файл
	url := "file://" + htmlPath
	gotoOpts := playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}
	if opts.Timeout > 0 {
		gotoOpts.Timeout = playwright.Float(opts.Timeout)
	}

	if _, err = page.Goto(url, gotoOpts); err != nil {
		return nil, "", err
	}

	// Ждем загрузки всех ресурсов
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, "", err
	}

	// Настраиваем параметры скриншота
	screenshotOpts := playwright.PageScreenshotOptions{
		FullPage:       playwright.Bool(opts.FullPage),
		OmitBackground: playwright.Bool(opts.OmitBackground),
	}

	// Определяем тип и content-type
	contentType := "image/png"

	switch opts.Type {
	case "jpeg", "jpg":
		screenshotType := playwright.ScreenshotTypeJpeg
		screenshotOpts.Type = screenshotType
		contentType = "image/jpeg"
		if opts.Quality != nil {
			screenshotOpts.Quality = playwright.Int(*opts.Quality)
		}
	default:
		screenshotType := playwright.ScreenshotTypePng
		screenshotOpts.Type = screenshotType
	}

	// Если указана область выделения
	if opts.Selections != nil {

		// Стандартный стиль, если не указан
		style := opts.SelectionStyle
		if style == nil {
			style = &SelectionStyle{
				BorderColor: "#FF0000",
				BorderWidth: 2,
				BorderStyle: "dashed",
				Opacity:     1.0,
			}
		}

		for i, selection := range opts.Selections {
			// Проверяем валидность координат
			if selection.Width <= 0 || selection.Height <= 0 {
				return nil, "", fmt.Errorf("invalid selection dimensions: width and height must be positive")
			}

			// JavaScript код для добавления прямоугольника выделения
			js := fmt.Sprintf(`
				(() => {
					const div = document.createElement('div');
					div.id = 'selection-rect-%d';
					div.style.position = 'absolute';
					div.style.left = '%dpx';
					div.style.top = '%dpx';
					div.style.width = '%dpx';
					div.style.height = '%dpx';
					div.style.border = '%dpx %s %s';
					div.style.opacity = '%f';
					div.style.boxSizing = 'border-box';
					div.style.zIndex = '2147483647';
					div.style.pointerEvents = 'none';
					document.body.appendChild(div);
				})()
			`,
				i,
				selection.X,
				selection.Y,
				selection.Width,
				selection.Height,
				style.BorderWidth,
				style.BorderStyle,
				style.BorderColor,
				style.Opacity,
			)

			// Выполняем JavaScript на странице
			if _, err := page.Evaluate(js); err != nil {
				return nil, "", fmt.Errorf("failed to draw selection rectangle: %w", err)
			}
		}
	}

	// Делаем скриншот в память
	bytes, err := page.Screenshot(screenshotOpts)
	if err != nil {
		return nil, "", err
	}

	return bytes, contentType, nil
}
