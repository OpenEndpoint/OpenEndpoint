package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// PoolConfig contains connection pool configuration
type PoolConfig struct {
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// PooledBackend is a connection pool wrapper for storage backends
type PooledBackend struct {
	backend  Backend
	pool     chan *pooledConn
	config   PoolConfig
	mu       sync.Mutex
	stats    PoolStats
}

// PoolStats contains pool statistics
type PoolStats struct {
	OpenConnections  int
	InUseConnections int
	IdleConnections  int
	WaitCount       int64
	WaitDuration     time.Duration
}

// pooledConn represents a pooled connection
type pooledConn struct {
	conn     interface{}
	created  time.Time
	lastUsed time.Time
}

// NewPooledBackend creates a new pooled storage backend
func NewPooledBackend(backend Backend, config PoolConfig) *PooledBackend {
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = DefaultPoolConfig().MaxOpenConns
	}
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = DefaultPoolConfig().MaxIdleConns
	}

	pool := &PooledBackend{
		backend: backend,
		pool:    make(chan *pooledConn, config.MaxIdleConns),
		config:  config,
	}

	// Pre-populate idle connections
	for i := 0; i < config.MaxIdleConns; i++ {
		pool.pool <- &pooledConn{
			created:  time.Now(),
			lastUsed: time.Now(),
		}
	}

	return pool
}

// Get acquires a connection from the pool
func (p *PooledBackend) Get(ctx context.Context) (*pooledConn, error) {
	p.mu.Lock()
	p.stats.OpenConnections++
	p.mu.Unlock()

	select {
	case conn := <-p.pool:
		p.mu.Lock()
		p.stats.IdleConnections--
		p.stats.InUseConnections++
		p.mu.Unlock()
		return conn, nil
	default:
		// No idle connections, wait for one
		select {
		case conn := <-p.pool:
			p.mu.Lock()
			p.stats.IdleConnections--
			p.stats.InUseConnections++
			p.mu.Unlock()
			return conn, nil
		case <-ctx.Done():
			p.mu.Lock()
			p.stats.WaitCount++
			p.mu.Unlock()
			return nil, ctx.Err()
		}
	}
}

// Put returns a connection to the pool
func (p *PooledBackend) Put(conn *pooledConn) {
	conn.lastUsed = time.Now()

	p.mu.Lock()
	p.stats.InUseConnections--
	p.stats.IdleConnections++
	p.mu.Unlock()

	select {
	case p.pool <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		p.mu.Lock()
		p.stats.IdleConnections--
		p.mu.Unlock()
	}
}

// Stats returns pool statistics
func (p *PooledBackend) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

// PutObject stores an object
func (p *PooledBackend) PutObject(ctx context.Context, bucket, key string, data io.Reader, opts PutOptions) (PutResult, error) {
	conn, err := p.Get(ctx)
	if err != nil {
		return PutResult{}, err
	}
	defer p.Put(conn)

	result, err := p.backend.PutObject(ctx, bucket, key, data, opts)
	if err != nil {
		p.mu.Lock()
		p.stats.WaitDuration += time.Since(conn.lastUsed)
		p.mu.Unlock()
	}
	return result, err
}

// GetObject retrieves an object
func (p *PooledBackend) GetObject(ctx context.Context, bucket, key string, opts GetOptions) (GetResult, error) {
	conn, err := p.Get(ctx)
	if err != nil {
		return GetResult{}, err
	}
	defer p.Put(conn)

	return p.backend.GetObject(ctx, bucket, key, opts)
}

// DeleteObject deletes an object
func (p *PooledBackend) DeleteObject(ctx context.Context, bucket, key string, opts DeleteOptions) error {
	conn, err := p.Get(ctx)
	if err != nil {
		return err
	}
	defer p.Put(conn)

	return p.backend.DeleteObject(ctx, bucket, key, opts)
}

// HeadObject returns object metadata without reading the body
func (p *PooledBackend) HeadObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	conn, err := p.Get(ctx)
	if err != nil {
		return ObjectInfo{}, err
	}
	defer p.Put(conn)

	return p.backend.HeadObject(ctx, bucket, key)
}

// ListObjects lists objects with prefix and delimiter
func (p *PooledBackend) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, string, error) {
	conn, err := p.Get(ctx)
	if err != nil {
		return nil, "", err
	}
	defer p.Put(conn)

	return p.backend.ListObjects(ctx, bucket, prefix, opts)
}

// Close closes the backend
func (p *PooledBackend) Close() error {
	close(p.pool)
	return p.backend.Close()
}

// BatchedBackend wraps a backend with request batching support
type BatchedBackend struct {
	backend   Backend
	batchCh  chan batchRequest
	mu       sync.Mutex
	workers  int
}

// batchRequest represents a batched request
type batchRequest struct {
	ctx     context.Context
	ops     []BatchOp
	results chan []BatchResult
	done    chan struct{}
}

// BatchOp represents a batched operation
type BatchOp struct {
	OpType  string // "get", "put", "delete"
	Bucket  string
	Key     string
	Data    io.Reader
	Opts    interface{}
}

// BatchResult represents the result of a batched operation
type BatchResult struct {
	Result interface{}
	Error  error
}

// NewBatchedBackend creates a new batched backend
func NewBatchedBackend(backend Backend, bufferSize int, workers int) *BatchedBackend {
	if workers <= 0 {
		workers = 4
	}

	b := &BatchedBackend{
		backend:  backend,
		batchCh:  make(chan batchRequest, bufferSize),
		workers:  workers,
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go b.worker()
	}

	return b
}

// worker processes batch requests
func (b *BatchedBackend) worker() {
	for {
		req, ok := <-b.batchCh
		if !ok {
			return
		}

		results := make([]BatchResult, len(req.ops))
		for i, op := range req.ops {
			switch op.OpType {
			case "get":
				result, err := b.backend.GetObject(req.ctx, op.Bucket, op.Key, GetOptions{})
				results[i] = BatchResult{Result: result, Error: err}
			case "head":
				result, err := b.backend.HeadObject(req.ctx, op.Bucket, op.Key)
				results[i] = BatchResult{Result: result, Error: err}
			case "delete":
				err := b.backend.DeleteObject(req.ctx, op.Bucket, op.Key, DeleteOptions{})
				results[i] = BatchResult{Error: err}
			default:
				results[i] = BatchResult{Error: errors.New("unknown operation type")}
			}
		}

		req.results <- results
		close(req.done)
	}
}

// ExecuteBatch executes multiple operations in a batch
func (b *BatchedBackend) ExecuteBatch(ctx context.Context, ops []BatchOp) ([]BatchResult, error) {
	req := batchRequest{
		ctx:     ctx,
		ops:     ops,
		results: make([]BatchResult, len(ops)),
		done:    make(chan struct{}),
	}

	select {
	case b.batchCh <- req:
		// Wait for results
		results := <-req.results
		return results, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// PutObject stores an object (delegates to backend)
func (b *BatchedBackend) PutObject(ctx context.Context, bucket, key string, data io.Reader, opts PutOptions) (PutResult, error) {
	return b.backend.PutObject(ctx, bucket, key, data, opts)
}

// GetObject retrieves an object (delegates to backend)
func (b *BatchedBackend) GetObject(ctx context.Context, bucket, key string, opts GetOptions) (GetResult, error) {
	return b.backend.GetObject(ctx, bucket, key, opts)
}

// DeleteObject deletes an object (delegates to backend)
func (b *BatchedBackend) DeleteObject(ctx context.Context, bucket, key string, opts DeleteOptions) error {
	return b.backend.DeleteObject(ctx, bucket, key, opts)
}

// HeadObject returns object metadata (delegates to backend)
func (b *BatchedBackend) HeadObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	return b.backend.HeadObject(ctx, bucket, key)
}

// ListObjects lists objects (delegates to backend)
func (b *BatchedBackend) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, string, error) {
	return b.backend.ListObjects(ctx, bucket, prefix, opts)
}

// Close closes the backend
func (b *BatchedBackend) Close() error {
	close(b.batchCh)
	return b.backend.Close()
}
