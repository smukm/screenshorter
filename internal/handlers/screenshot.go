package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"screenshoter/internal/service"
	"time"
)

// метрики для prometeus
var (
	activeWorkersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "screenshot_service_active_workers",
		Help: "Current number of active screenshot workers",
	})

	totalRequestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "screenshot_service_requests_total",
			Help: "Total number of screenshot requests",
		},
		[]string{"status"},
	)

	requestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "screenshot_service_request_duration_seconds",
			Help:    "Duration of screenshot requests",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"browser"},
	)
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(activeWorkersGauge)
	prometheus.MustRegister(totalRequestsCounter)
	prometheus.MustRegister(requestDurationHistogram)
}

func (h *Handler) Make(ctx *gin.Context) {

	startTime := time.Now()

	// Определяем браузер
	defaultBrowser := service.BrowserChromium
	browser := ctx.PostForm("browser")
	switch browser {
	case "firefox":
		defaultBrowser = service.BrowserFirefox
	case "webkit":
		defaultBrowser = service.BrowserWebkit
	}

	defer func() {
		if ctx.Writer.Status() >= 200 && ctx.Writer.Status() < 300 {
			// Записываем продолжительность запроса
			requestDurationHistogram.WithLabelValues(browser).Observe(time.Since(startTime).Seconds())
		}
	}()

	html := ctx.PostForm("html")
	if html == "" {
		totalRequestsCounter.WithLabelValues("400").Inc()
		newErrorResponse(ctx, http.StatusBadRequest, "html content is required")
		return
	}

	// Пытаемся занять слот в пуле воркеров
	select {
	case h.workerPool <- struct{}{}: // Получили слот
		activeWorkersGauge.Inc() // Увеличиваем счетчик активных воркеров
		defer func() {
			<-h.workerPool
			activeWorkersGauge.Dec() // Уменьшаем счетчик при освобождении
		}()
	case <-ctx.Request.Context().Done():
		totalRequestsCounter.WithLabelValues("429").Inc()
		newErrorResponse(ctx, http.StatusTooManyRequests, "request cancelled by client")
		return
	default:
		totalRequestsCounter.WithLabelValues("429").Inc()
		newErrorResponse(ctx, http.StatusTooManyRequests, "server busy, try again later")
		return
	}

	// Настройки скриншота
	opts := service.ScreenshotOptions{
		Browser:        defaultBrowser,
		Type:           h.cfg.Type,
		Quality:        nil,
		FullPage:       true,
		OmitBackground: false,
		/*Viewport: (*struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		})(&struct {
			Width  int
			Height int
		}{Width: 1200, Height: 800}),*/
		Timeout: 5000,
	}

	// Канал для результата
	resultChan := make(chan struct {
		bytes       []byte
		contentType string
		err         error
	}, 1)

	// Запускаем создание скриншота в горутине
	go func() {
		bytes, contentType, err := h.service.Screenshot.Make(html, opts)
		resultChan <- struct {
			bytes       []byte
			contentType string
			err         error
		}{bytes, contentType, err}
	}()

	// Ждем результат с таймаутом
	select {
	case result := <-resultChan:
		if result.err != nil {
			totalRequestsCounter.WithLabelValues("500").Inc()
			newErrorResponse(ctx, http.StatusInternalServerError, result.err.Error())
			return
		}
		totalRequestsCounter.WithLabelValues("200").Inc()
		ctx.Data(http.StatusOK, result.contentType, result.bytes)
	case <-ctx.Request.Context().Done():
		totalRequestsCounter.WithLabelValues("499").Inc()
		newErrorResponse(ctx, http.StatusRequestTimeout, "request timeout")
		return
	case <-time.After(20 * time.Second): // Таймаут на выполнение
		totalRequestsCounter.WithLabelValues("504").Inc()
		newErrorResponse(ctx, http.StatusRequestTimeout, "screenshot generation timeout")
		return
	}

}

// Дополнительные кастомные метрики
func (h *Handler) MetricsHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"active_workers": len(h.workerPool),
		"max_workers":    h.cfg.MaxWorkers,
		"queue_size":     h.cfg.MaxWorkers - len(h.workerPool),
	})
}
