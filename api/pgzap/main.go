package pgzap

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap/zapcore"
)

type Core struct {
	enc       zapcore.Encoder
	level     zapcore.LevelEnabler
	db        *sql.DB
	ch        chan []byte
	stop      chan struct{}
	wg        sync.WaitGroup
	batchSize int
	maxWait   time.Duration
	reqKeys   []string
}

type Option func(*Core)

func WithBatchSize(n int) Option {
	return func(c *Core) {
		if n > 0 {
			c.batchSize = n
		}
	}
}
func WithMaxWait(d time.Duration) Option {
	return func(c *Core) {
		if d > 0 {
			c.maxWait = d
		}
	}
}
func WithBuffer(n int) Option {
	return func(c *Core) {
		if n > 0 {
			c.ch = make(chan []byte, n)
		}
	}
}

func WithRequestIDKeys(keys ...string) Option {
	return func(c *Core) {
		if len(keys) > 0 {
			c.reqKeys = keys
		}
	}
}

func New(db *sql.DB, level zapcore.LevelEnabler, opts ...Option) (*Core, func(context.Context) error) {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		NameKey:        "logger",
		StacktraceKey:  "stack",
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	c := &Core{
		enc:       zapcore.NewJSONEncoder(encCfg),
		level:     level,
		db:        db,
		ch:        make(chan []byte, 10_000),
		stop:      make(chan struct{}),
		batchSize: 1000,
		maxWait:   2 * time.Second,
		reqKeys:   []string{"request_id", "req_id", "X-Request-ID", "X-Correlation-ID", "requestId"},
	}
	for _, o := range opts {
		o(c)
	}

	c.wg.Add(1)
	go c.loop()

	closeFn := func(ctx context.Context) error {
		close(c.stop)
		done := make(chan struct{})
		go func() { c.wg.Wait(); close(done) }()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
	return c, closeFn
}

func (c *Core) Enabled(lvl zapcore.Level) bool { return c.level.Enabled(lvl) }

func (c *Core) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.enc = c.enc.Clone()
	for _, f := range fields {
		f.AddTo(clone.enc)
	}
	return &clone
}

func (c *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *Core) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	b := append([]byte(nil), buf.Bytes()...)
	buf.Free()

	select {
	case c.ch <- b:
	default:
		log.Printf("pgzap: dropping log (buffer full)")
	}
	return nil
}

func (c *Core) Sync() error { return nil }

func (c *Core) loop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.maxWait)
	defer ticker.Stop()

	batch := make([][]byte, 0, c.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		tx, err := c.db.Begin()
		if err != nil {
			log.Printf("pgzap: begin err: %v", err)
			batch = batch[:0]
			return
		}

		stmt, err := tx.Prepare(pq.CopyIn("logs", "req_id", "raw"))
		if err != nil {
			log.Printf("pgzap: prepare err: %v", err)
			_ = tx.Rollback()
			batch = batch[:0]
			return
		}

		for _, line := range batch {
			var tmp map[string]any
			if json.Unmarshal(line, &tmp) != nil {
				continue // ignore lignes non-JSON
			}
			// Cherche request_id dans plusieurs clés possibles
			var reqID string
			for _, k := range c.reqKeys {
				if v, ok := tmp[k]; ok {
					if s, ok := v.(string); ok && s != "" {
						reqID = s
						break
					}
				}
			}
			if _, err := stmt.Exec(reqID, string(line)); err != nil {
				log.Printf("pgzap: exec err: %v", err)
			}
		}

		_, _ = stmt.Exec()
		_ = stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Printf("pgzap: commit err: %v", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-c.stop:
			flush()
			return
		case b := <-c.ch:
			batch = append(batch, b)
			if len(batch) >= c.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
