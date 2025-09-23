package db

import (
	"context"
	"fmt"
	"strconv"

	models "github.com/Valentin-Makurin/metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Database struct {
	pool *pgxpool.Pool
	log  *zap.SugaredLogger
}

func NewDatabase(connString string, logger *zap.SugaredLogger) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), connString)
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
	return db.pool.Ping(context.Background())
}

func (db *Database) SetVal(key string, mtr models.Metrics) {
	res, err := db.pool.Exec(context.Background(), models.Upsert, &key, &mtr.MType, mtr.Delta, mtr.Value)
	if err != nil {
		db.log.Error("Failed to run upsert: %v", err)
	}

	count := res.RowsAffected()
	fmt.Printf("effect - %d", count)
}

func (db *Database) AddVal(key string, mtr models.Metrics) {

	var deltaOld int64
	err := db.pool.QueryRow(context.Background(), models.SelectDelta, &key, &mtr.MType).Scan(&deltaOld)
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
	err := db.pool.QueryRow(context.Background(), models.SelectRow, &key).Scan(&res.ID, &res.MType, &res.Delta, &res.Value)
	if err != nil {
		db.log.Error("Failed to run select row: %v", err)
	}

	return res
}

func (db *Database) GetAllVal() map[string]string {

	rows, err := db.pool.Query(context.Background(), models.SelectAll)
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

func (db *Database) RunMigrations() error {
	_, err := db.pool.Exec(context.Background(), models.Migration)
	if err != nil {
		db.log.Error("Failed to run migration: %v", err)
		return err
	}
	return nil
}
