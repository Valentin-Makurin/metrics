package handler

import (
	models "github.com/Valentin-Makurin/metrics/internal/model"
)

type MockStorage struct {
	SetValCalled      bool
	SetValArgs        []interface{}
	AddValCalled      bool
	AddValArgs        []interface{}
	GetValCalled      bool
	GetValKey         string
	GetValResult      models.Metrics
	GetAllValCalled   bool
	GetAllValResult   map[string]string
	UpsertBatchCalled bool
	UpsertBatchError  error
	PingCalled        bool
	PingError         error
	SetValCustom      func(key string, mtr models.Metrics)
	AddValCustom      func(key string, mtr models.Metrics)
	GetValCustom      func(key string) models.Metrics
	GetAllValCustom   func() map[string]string
	UpsertBatchCustom func(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error
	PingCustom        func() error
}

func (m *MockStorage) SetVal(key string, mtr models.Metrics) {
	m.SetValCalled = true
	m.SetValArgs = []interface{}{key, mtr}
	if m.SetValCustom != nil {
		m.SetValCustom(key, mtr)
	}
}

func (m *MockStorage) AddVal(key string, mtr models.Metrics) {
	m.AddValCalled = true
	m.AddValArgs = []interface{}{key, mtr}
	if m.AddValCustom != nil {
		m.AddValCustom(key, mtr)
	}
}

func (m *MockStorage) GetVal(key string) models.Metrics {
	m.GetValCalled = true
	m.GetValKey = key
	if m.GetValCustom != nil {
		return m.GetValCustom(key)
	}
	return m.GetValResult
}

func (m *MockStorage) GetAllVal() map[string]string {
	m.GetAllValCalled = true
	if m.GetAllValCustom != nil {
		return m.GetAllValCustom()
	}
	return m.GetAllValResult
}

func (m *MockStorage) UpsertBatch(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error {
	m.UpsertBatchCalled = true
	if m.UpsertBatchCustom != nil {
		return m.UpsertBatchCustom(GaugeMtr, CntMtr)
	}
	return m.UpsertBatchError
}

func (m *MockStorage) Ping() error {
	m.PingCalled = true
	if m.PingCustom != nil {
		return m.PingCustom()
	}
	return m.PingError
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
