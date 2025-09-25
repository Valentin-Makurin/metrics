package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"go.uber.org/zap"
)

type Storage interface {
	SetVal(key string, mtr models.Metrics)
	AddVal(key string, mtr models.Metrics)
	GetVal(key string) models.Metrics
	GetAllVal() map[string]string
	Ping() error
	UpsertBatch(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error
}
type MtrHandler struct {
	storage Storage
	logger  *zap.SugaredLogger
}

func NewMtrHandler(stor Storage, logger *zap.SugaredLogger) *MtrHandler {
	return &MtrHandler{
		storage: stor,
		logger:  logger,
	}
}

func (h *MtrHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 || pathParts[1] != "update" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValueStr := pathParts[4]

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !TypeCheck(metricType) {
		w.WriteHeader(http.StatusBadRequest)
	}

	switch metricType {
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.SetVal(metricName, models.Metrics{MType: models.Gauge, Value: &value})

	case models.Counter:
		value, err := strconv.ParseInt(metricValueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.AddVal(metricName, models.Metrics{MType: models.Counter, Delta: &value})
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MtrHandler) HandlePostUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !TypeCheck(metric.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.SetVal(metric.ID, metric)

	case models.Counter:
		if metric.Delta == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.AddVal(metric.ID, metric)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MtrHandler) HandlePostUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var rawMetrics []models.Metrics
	var validMetricsGauge []models.Metrics
	validMetricsCounter := make(map[string]models.Metrics)

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rawMetrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, val := range rawMetrics {
		if val.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if !TypeCheck(val.MType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch val.MType {
		case models.Gauge:
			if val.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// h.storage.SetVal(metric.ID, metric)
			validMetricsGauge = append(validMetricsGauge, val)

		case models.Counter:
			if val.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mtr, ok := validMetricsCounter[val.ID]
			if ok {
				*val.Delta += *mtr.Delta
			}
			validMetricsCounter[val.ID] = val
			// h.storage.AddVal(metric.ID, metric)
		}

		// validMetrics = append(validMetrics, val)
	}

	// switch metric.MType {
	// case models.Gauge:
	// 	if metric.Value == nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	h.storage.SetVal(metric.ID, metric)

	// case models.Counter:
	// 	if metric.Delta == nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	h.storage.AddVal(metric.ID, metric)
	// }

	err := h.storage.UpsertBatch(validMetricsGauge, validMetricsCounter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func TypeCheck(metricType string) bool {
	return metricType == models.Gauge || metricType == models.Counter
}

func (h *MtrHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 || pathParts[1] != "value" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricType := pathParts[2]
	metricName := pathParts[3]
	if !TypeCheck(metricType) {
		w.WriteHeader(http.StatusBadRequest)
	}

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	res := h.storage.GetVal(metricName)
	if res.MType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	valStr := ""
	switch res.MType {
	case models.Gauge:
		valStr = strconv.FormatFloat(*res.Value, 'f', -1, 64)
		if strings.Contains(valStr, ".") {
			valStr = strings.TrimRight(valStr, "0")
		}
	case models.Counter:
		valStr = strconv.FormatInt(*res.Delta, 10)
	}
	_, err := w.Write([]byte(valStr))
	if err != nil {
		h.logger.Error("Failed to write response", err)
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MtrHandler) HandleGetValue(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var metric models.Metrics
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !TypeCheck(metric.MType) {
		w.WriteHeader(http.StatusBadRequest)
	}

	if metric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	res := h.storage.GetVal(metric.ID)
	if res.MType == "" || res.MType != metric.MType {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metric.Value = res.Value
	metric.Delta = res.Delta

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("Error encoding JSON response", err)
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MtrHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	metrics := h.storage.GetAllVal()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := w.Write([]byte(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Метрики</title>
            <style>
                table { border-collapse: collapse; width: 100%; }
                th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
                th { background-color: #f2f2f2; }
                tr:nth-child(even) { background-color: #f9f9f9; }
            </style>
        </head>
        <body>
            <h1>Список метрик</h1>
            <table>
                <thead>
                    <tr>
                        <th>Имя метрики</th>
                        <th>Значение</th>
                    </tr>
                </thead>
                <tbody>
    `))
	if err != nil {
		h.logger.Error("Filed to write html top", err)
	}

	for name, value := range metrics {
		_, err = fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", html.EscapeString(name), html.EscapeString(value))
		if err != nil {
			h.logger.Error("Filed to write html body", err)
		}

	}

	_, err = w.Write([]byte(`
                </tbody>
            </table>
        </body>
        </html>
    `))
	if err != nil {
		h.logger.Error("Filed to write html footer", err)
	}

}

func (h *MtrHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	err := h.storage.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

}
