package service

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"screenshoter/pkg/logger"

	"github.com/playwright-community/playwright-go"
)

type BrowserType string

const (
	BrowserChromium BrowserType = "chromium"
	BrowserFirefox  BrowserType = "firefox"
	BrowserWebkit   BrowserType = "webkit"
)

type Playwright struct {
	pw  *playwright.Playwright
	lgr *logger.Logger
}

func NewPlaywright(lgr *logger.Logger) (*Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not launch playwright: %w", err)
	}
	return &Playwright{pw: pw, lgr: lgr}, nil
}

// Close освобождает ресурсы
func (p *Playwright) Close() error {
	return p.pw.Stop()
}

// Make формирует скриншот из html
func (p *Playwright) Make(html string, opts ScreenshotOptions) ([]byte, string, error) {
	if html == "" {
		return nil, "", fmt.Errorf("html content cannot be empty")
	}
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
	defer func() {
		if closeErr := browser.Close(); closeErr != nil {
			p.lgr.Warn().Msgf("failed to close browser: %v", closeErr)
		}
	}()

	// Создаем временный файл со случайным именем
	htmlPath, err := p.createTempHTML(html)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if removeErr := os.Remove(htmlPath); removeErr != nil {
			p.lgr.Warn().Msgf("failed to remove temp file %s: %v", htmlPath, removeErr)
		}
	}()

	page, err := browser.NewPage()
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if closeErr := page.Close(); closeErr != nil {
			p.lgr.Warn().Msgf("failed to close page: %v", closeErr)
		}
	}()

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

		// Получаем текущую прокрутку страницы
		scrollJS := `({ scrollX: window.scrollX, scrollY: window.scrollY })`
		scrollResult, err := page.Evaluate(scrollJS)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get page scroll position: %w", err)
		}

		// Safe type conversion
		scrollData, ok := scrollResult.(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("unexpected scroll position format")
		}

		// Helper function to safely convert scroll values to int
		getScrollValue := func(val interface{}) int {
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case int32:
				return int(v)
			case int64:
				return int(v)
			default:
				return 0
			}
		}

		currentScrollX := getScrollValue(scrollData["scrollX"])
		currentScrollY := getScrollValue(scrollData["scrollY"])

		for i, selection := range opts.Selections {
			// Проверяем валидность координат
			if selection.Width <= 0 || selection.Height <= 0 {
				return nil, "", fmt.Errorf("invalid selection dimensions: width and height must be positive")
			}

			// Рассчитываем позицию с учетом прокрутки
			effectiveX := selection.X
			effectiveY := selection.Y
			// Если указаны scrollX/Y, используем их, иначе текущую прокрутку
			if selection.ScrollX != 0 {
				effectiveX += selection.ScrollX
			} else {
				effectiveX += currentScrollX
			}

			if selection.ScrollY != 0 {
				effectiveY += selection.ScrollY
			} else {
				effectiveY += currentScrollY
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
				effectiveX,
				effectiveY,
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

// createTempHtml создает файл с контентом со случайным именем
func (p *Playwright) createTempHTML(content string) (string, error) {
	tmpDir := os.TempDir()

	// Создаем случайное имя файла
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("failed to generate random filename: %w", err)
	}
	filename := fmt.Sprintf("screenshot_%x.html", randBytes)
	htmlPath := filepath.Join(tmpDir, filename)
	return htmlPath, os.WriteFile(htmlPath, []byte(content), 0644)
}
