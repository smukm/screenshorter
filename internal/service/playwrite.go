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

	// Делаем скриншот в память
	bytes, err := page.Screenshot(screenshotOpts)
	if err != nil {
		return nil, "", err
	}

	return bytes, contentType, nil
}
