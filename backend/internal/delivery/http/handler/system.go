package handler

import (
	"context"
	"net/http"

	domainsystem "github.com/PhamVanPhuc2k2/manage-company/internal/domain/system"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/apperror"
	"github.com/PhamVanPhuc2k2/manage-company/pkg/httpx"
	"github.com/go-chi/chi/v5/middleware"
)

type PingUsecase interface {
	Ping(ctx context.Context, requestID string) (domainsystem.HealthSnapshot, error)
}

type SystemHanler struct {
	uc      PingUsecase
	version string
	commit  string
}

func NewSystemHandler(uc PingUsecase, version, commit string) *SystemHanler {
	return &SystemHanler{
		uc:      uc,
		version: version,
		commit:  commit,
	}
}

func (h *SystemHanler) Ping(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	snapshot, err := h.uc.Ping(r.Context(), requestID)
	if err != nil {
		Error(w, err, requestID)
		return
	}

	httpx.OK(w, map[string]any{
		"database_time": snapshot.DatabaseTime,
		"redis_ok":      snapshot.RedisOk,
		"job_queued":    snapshot.JobQueued,
		"request_id":    snapshot.RequestID,
		"version":       h.version,
		"commit":        h.commit,
	})
}

func Error(w http.ResponseWriter, err error, requestID string) {
	status, appErr := apperror.HTTPStatus(err)

	message := appErr.Message

	if status == http.StatusInternalServerError {
		message = "Đã có lỗi xảy ra, vui lòng thử lại"
	}

	httpx.JSON(w, status, httpx.Envelope{Error: &httpx.ErrorBody{
		Code:      string(appErr.Kind),
		Message:   message,
		Details:   appErr.Details,
		RequestID: requestID,
	}})
}
