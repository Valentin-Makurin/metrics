package handler

import (
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

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
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
		h.storage.SetVal(metricName, models.Metrics{MType: models.Gauge, Delta: &value})
	}
	w.WriteHeader(http.StatusOK)
}

func TypeCheck(metricType string) bool {
	return metricType == models.Gauge || metricType == models.Counter
}
