package system

import (
	"context"

	domainsystem "github.com/PhamVanPhuc2k2/manage-company/internal/domain/system"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/apperror"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/logger"
)

type PingUsecase struct {
	clock     domainsystem.Clock
	cache     domainsystem.Cache
	publisher domainsystem.JobPublisher
}

func NewPingUsecase(clock domainsystem.Clock, cache domainsystem.Cache, publisher domainsystem.JobPublisher) *PingUsecase {
	return &PingUsecase{
		clock:     clock,
		cache:     cache,
		publisher: publisher,
	}
}

func (u *PingUsecase) Ping(ctx context.Context, requestID string) (domainsystem.HealthSnapshot, error) {
	log := logger.FromContext(ctx)

	dbTime, err := u.clock.Now(ctx)
	if err != nil {
		return domainsystem.HealthSnapshot{}, apperror.Wrap(apperror.KindInternal, "không đọc được giờ từ database", err)
	}

	redisOK := true
	if err := u.cache.Ping(ctx); err != nil {
		log.Warn().Err(err).Msg("redis không phản hồi")
		redisOK = false
	}

	jobQueued := true
	if err := u.publisher.Publish(ctx, domainsystem.Job{
		Name:      domainsystem.JobSystemPing,
		RequestID: requestID,
		Payload:   map[string]any{"source": "api"},
	}); err != nil {
		log.Warn().Err(err).Msg("không đẩy được job lên hàng đợi")
		jobQueued = false
	}

	return domainsystem.HealthSnapshot{
		DatabaseTime: dbTime,
		RedisOk:      redisOK,
		JobQueued:    jobQueued,
		RequestID:    requestID,
	}, nil
}
