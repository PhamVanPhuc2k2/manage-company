package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	ExchangeJobs     = "manage.jobs"
	ExchangeJobsDLX  = "manage.jobs.dlx"
	QueueJobs        = "manage.jobs.queue"
	QueueJobsDead    = "manage.jobs.dead"
	ExchangeRealtime = "manage.realtime"
)

var ErrNotConnected = errors.New("rabbitmq không có kết nối")

type Client struct {
	url       string
	log       zerolog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	conn      *amqp.Connection
	ch        *amqp.Channel
	closed    bool
	boundJobs map[string]struct{}
}

func New(ctx context.Context, url string, log zerolog.Logger) (*Client, error) {
	clientCtx, cancel := context.WithCancel(ctx)

	c := &Client{
		url:       url,
		log:       log,
		ctx:       clientCtx,
		cancel:    cancel,
		boundJobs: make(map[string]struct{}),
	}
	if err := c.connect(clientCtx); err != nil {
		cancel()
		return nil, err
	}
	go c.watchAndReconnect()
	return c, nil
}

func (c *Client) connect(ctx context.Context) error {
	const maxAttempts = 10
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := amqp.Dial(c.url)
		if err == nil {
			ch, chErr := conn.Channel()
			if chErr == nil {
				c.mu.Lock()
				c.conn, c.ch = conn, ch
				c.mu.Unlock()

				if err := c.declareTopology(); err != nil {
					return err
				}
				c.log.Info().Msg("đã kết nối rabbitmq")
				return nil
			}
			_ = conn.Close()
			lastErr = chErr
		} else {
			lastErr = err
		}

		wait := time.Duration(attempt) * time.Second
		c.log.Warn().Err(err).Int("attempt", attempt).Dur("retry_in", wait).Msg("chưa kết nối được rabbitmq, thử lại")

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("không kết nối được rabbitmq sau %d lần: %w", maxAttempts, lastErr)
}

func (c *Client) declareTopology() error {
	ch := c.Channel()
	if ch == nil {
		return ErrNotConnected
	}

	if err := ch.ExchangeDeclare(ExchangeJobs, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("khai báo exchange %s: %w", ExchangeJobs, err)
	}
	if err := ch.ExchangeDeclare(ExchangeJobsDLX, "fanout", true, false, false, false, nil); err != nil {
		return fmt.Errorf("khái báo exchange %s: %w", ExchangeJobsDLX, err)
	}
	if err := ch.ExchangeDeclare(ExchangeRealtime, "fanout", true, false, false, false, nil); err != nil {
		return fmt.Errorf("khai báo exchange %s: %w", ExchangeRealtime, err)
	}

	if _, err := ch.QueueDeclare(QueueJobs, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": ExchangeJobsDLX,
	}); err != nil {
		return fmt.Errorf("khai báo queue %s: %w", QueueJobs, err)
	}
	if _, err := ch.QueueDeclare(QueueJobsDead, true, false, false, false, nil); err != nil {
		return fmt.Errorf("khái báo queue %s: %w", QueueJobsDead, err)
	}
	if err := ch.QueueBind(QueueJobsDead, "", ExchangeJobsDLX, false, nil); err != nil {
		return fmt.Errorf("bind queue chết: %w", err)
	}

	c.mu.RLock()
	jobs := make([]string, 0, len(c.boundJobs))
	for name := range c.boundJobs {
		jobs = append(jobs, name)
	}
	c.mu.RUnlock()

	for _, name := range jobs {
		if err := ch.QueueBind(QueueJobs, name, ExchangeJobs, false, nil); err != nil {
			return fmt.Errorf("bind lại job %s: %w", name, err)
		}
	}

	return nil
}

func (c *Client) BindJob(jobName string) error {
	ch := c.Channel()
	if ch == nil {
		return ErrNotConnected
	}
	if err := ch.QueueBind(QueueJobs, jobName, ExchangeJobs, false, nil); err != nil {
		return err
	}

	c.mu.Lock()
	c.boundJobs[jobName] = struct{}{}
	c.mu.Unlock()
	return nil
}

func (c *Client) watchAndReconnect() {
	for {
		c.mu.RLock()
		conn := c.conn
		closed := c.closed
		c.mu.RUnlock()

		if conn == nil || closed {
			return
		}

		reason, ok := <-conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			return
		}

		c.mu.RLock()
		closed = c.closed
		c.mu.RUnlock()
		if closed {
			return
		}

		c.log.Error().Str("reason", reason.Error()).Msg("rabbitmq mất kết nối, đang kết nối lại")
		if err := c.connect(c.ctx); err != nil {
			c.log.Error().Err(err).Msg("kết nối lại rabbitmq thất bại")
			return
		}
	}
}

func (c *Client) Channel() *amqp.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ch
}

func (c *Client) NewChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return nil, ErrNotConnected
	}
	return conn.Channel()
}

func (c *Client) HealthCheck(context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil || c.conn.IsClosed() {
		return ErrNotConnected
	}
	return nil
}

func (c *Client) Close() error {
	c.cancel()

	c.mu.Lock()
	c.closed = true
	ch, conn := c.ch, c.conn
	c.mu.Unlock()

	if ch != nil {
		_ = ch.Close()
	}
	if conn != nil {
		return conn.Close()
	}
	return nil
}
