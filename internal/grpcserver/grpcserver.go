package grpcserver

import (
	"context"
	"strings"

	"github.com/Valentin-Makurin/metrics/internal/common"
	models "github.com/Valentin-Makurin/metrics/internal/model"
	pb "github.com/Valentin-Makurin/metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CIDRChecker interface {
	CheckIPGRPC(ipStr string) bool
}

// BatchRepository определяет интерфейс для пакетных операций с метриками.
type BatchRepository interface {
	UpsertBatch(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error
}

// Storage объединяет все интерфейсы хранилища метрик.
type Storage interface {
	BatchRepository
}

type Server struct {
	pb.UnimplementedMetricsServer
	storage Storage
}

func NewServer(stor Storage) *Server {
	return &Server{
		storage: stor,
	}
}

func (s *Server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {

	validMetricsGauge := make([]models.Metrics, 0, 60)
	validMetricsCounter := make(map[string]models.Metrics)

	for _, val := range req.GetMetrics() {
		id := val.GetId()
		if id == "" {
			return nil, status.Errorf(codes.Internal, "missing IP")
		}
		mType := strings.ToLower(val.GetType().String())
		if !common.TypeCheck(mType) {
			return nil, status.Errorf(codes.Internal, "unregistred Type")
		}

		switch mType {
		case models.Gauge:
			value := val.GetValue()
			validMetricsGauge = append(validMetricsGauge, models.Metrics{ID: id, MType: mType, Value: &value})

		case models.Counter:
			delta := val.GetDelta()
			mtr, ok := validMetricsCounter[id]
			if ok {
				delta += *mtr.Delta
			}
			validMetricsCounter[id] = models.Metrics{ID: id, MType: mType, Delta: &delta}
		}

	}

	err := s.storage.UpsertBatch(validMetricsGauge, validMetricsCounter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "faile to save IP")
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func NewUnaryInterceptor(checker CIDRChecker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("X-Real-IP")
			if len(values) > 0 {
				IP := values[0]
				if !checker.CheckIPGRPC(IP) {
					return nil, status.Errorf(codes.PermissionDenied, "unregistered IP: %s", IP)
				}
			} else {
				return nil, status.Errorf(codes.Unauthenticated, "missing IP")
			}
		} else {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		return handler(ctx, req)
	}
}
