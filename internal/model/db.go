package models

const (
	Migration = `CREATE SCHEMA IF NOT EXISTS mvv;
	CREATE TABLE IF NOT EXISTS mvv.metrics (
	id TEXT NOT NULL ,
	mtype TEXT NOT NULL ,
	delta BIGINT,
	value DOUBLE PRECISION,
	PRIMARY KEY (id, mtype));`

	Upsert = `
		INSERT INTO mvv.metrics (id, mtype, delta, value)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id, mtype) 
        DO UPDATE SET 
            delta = EXCLUDED.delta,
            value = EXCLUDED.value
        WHERE mvv.metrics.id = EXCLUDED.id 
		AND mvv.metrics.mtype = EXCLUDED.mtype`

	SelectDelta = `SELECT delta
        FROM mvv.metrics 
        WHERE id = $1 
		AND mtype = $2`

	SelectRow = `SELECT id, mtype, delta, value
        FROM mvv.metrics  
        WHERE id = $1`

	SelectAll = `SELECT id, mtype, delta, value FROM mvv.metrics `
)
