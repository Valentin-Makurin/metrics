package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	timeout      time.Duration = 5 * time.Second
	extraTimeout time.Duration = 40 * time.Second
)

type Database struct {
	pool *pgxpool.Pool
	log  *zap.SugaredLogger
}

func NewDatabase(ctx context.Context, connString string, logger *zap.SugaredLogger) (*Database, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать подключение: %w", err)
	}

	return &Database{
		pool: pool,
		log:  logger}, nil
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.pool.Ping(ctx)
}

func (db *Database) SetVal(key string, mtr models.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := db.pool.Exec(ctx, models.Upsert, &key, &mtr.MType, mtr.Delta, mtr.Value)
	if err != nil {
		db.log.Error("Failed to run upsert: %v", err)
	}

	count := res.RowsAffected()
	fmt.Printf("effect - %d", count)
}

func (db *Database) AddVal(key string, mtr models.Metrics) {
	var deltaOld int64

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := db.pool.QueryRow(ctx, models.SelectDelta, &key, &mtr.MType).Scan(&deltaOld)
	if err != nil {
		db.log.Error("Failed to run select delsta: %v", err)
	}

	if deltaOld != 0 {
		*mtr.Delta += deltaOld
	}
	db.SetVal(key, mtr)
}

func (db *Database) GetVal(key string) models.Metrics {
	var res models.Metrics
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := db.pool.QueryRow(ctx, models.SelectRow, &key).Scan(&res.ID, &res.MType, &res.Delta, &res.Value)
	if err != nil {
		db.log.Error("Failed to run select row: %v", err)
	}

	return res
}

func (db *Database) GetAllVal() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rows, err := db.pool.Query(ctx, models.SelectAll)
	if err != nil {
		db.log.Error("Failed to run select all: %v", err)
	}

	res := make(map[string]string)
	for rows.Next() {
		tmpMetric := models.Metrics{}
		err := rows.Scan(
			&tmpMetric.ID,
			&tmpMetric.MType,
			&tmpMetric.Delta,
			&tmpMetric.Value,
		)
		if err != nil {
			db.log.Error("Failed to scan metrics: %v", err)
			return res
		}
		str := ""
		if tmpMetric.Value != nil {
			str = fmt.Sprintf("%.4f", *tmpMetric.Value)
		} else if tmpMetric.Delta != nil {
			str = strconv.FormatInt(*tmpMetric.Delta, 10)
		}
		res[tmpMetric.ID] = str
	}
	if err = rows.Err(); err != nil {
		db.log.Error("Error during rows iteration: %v", err)
	}

	return res
}

func (db *Database) RunMigrations(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := db.pool.Exec(ctx, models.Migration)
	if err != nil {
		db.log.Error("Failed to run migration: %v", err)
		return err
	}
	return nil
}

func (db *Database) UpsertBatch(GaugeMtr []models.Metrics, CntMtr map[string]models.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), extraTimeout)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, val := range GaugeMtr {
		_, err = tx.Exec(ctx, models.Upsert, &val.ID, &val.MType, val.Delta, val.Value)
		if err != nil {
			db.log.Error("Failed to run upsert: %v", err)
		}
	}

	for _, val := range CntMtr {
		var deltaOld int64
		err := db.pool.QueryRow(ctx, models.SelectDelta, &val.ID, &val.MType).Scan(&deltaOld)
		if err != nil {
			db.log.Error("Failed to run select delsta: %v", err)
		}
		if deltaOld != 0 {
			*val.Delta += deltaOld
		}
		_, err = db.pool.Exec(ctx, models.Upsert, &val.ID, &val.MType, val.Delta, val.Value)
		if err != nil {
			db.log.Error("Failed to run upsert: %v", err)
		}
	}

	tx.Commit(ctx)
	return nil
}

func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
			return true
		}
	}

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.InsufficientResources,
			pgerrcode.DiskFull,
			pgerrcode.OutOfMemory,
			pgerrcode.TooManyConnections,
			pgerrcode.QueryCanceled:
			return true
		}
	}

	return false
}
