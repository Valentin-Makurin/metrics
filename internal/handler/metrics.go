package handler

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Valentin-Makurin/metrics/internal/db"
	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type MtrHandler struct {
	storage *db.MtrStorage
}

func NewMtrHandler(stor *db.MtrStorage) *MtrHandler {
	return &MtrHandler{
		storage: stor,
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
		log.Printf("Failed to write response")
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MtrHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	metrics := h.storage.GetAllVal()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, _ = w.Write([]byte(`
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

	for name, value := range metrics {
		_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", html.EscapeString(name), html.EscapeString(value))
	}

	_, _ = w.Write([]byte(`
                </tbody>
            </table>
        </body>
        </html>
    `))
}
