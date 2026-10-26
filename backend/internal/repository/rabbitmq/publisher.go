package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainsystem "github.com/PhamVanPhuc2k2/manage-company/internal/domain/system"
	mq "github.com/PhamVanPhuc2k2/manage-company/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type JobPublisher struct {
	client *mq.Client
}

func NewJobPublisher(client *mq.Client) *JobPublisher {
	return &JobPublisher{
		client: client,
	}
}

func (p *JobPublisher) Publish(ctx context.Context, job domainsystem.Job) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("mã hóa job: %w", err)
	}

	ch := p.client.Channel()
	if ch == nil {
		return mq.ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(
		ctx,
		mq.ExchangeJobs,
		job.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			MessageId:    job.RequestID,
			Body:         body,
		},
	)
}
