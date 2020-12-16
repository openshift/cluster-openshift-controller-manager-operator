package healthz

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/apiserver/pkg/server/healthz"
)

type livenessChecker struct {
	StopCh <-chan struct{}
}

// NewLivenessChecker creates a health checker
func NewLivenessChecker(ctx context.Context) healthz.HealthChecker {
	return &livenessChecker{
		StopCh: ctx.Done(),
	}
}

func (*livenessChecker) Name() string {
	return "livez"
}

func (l *livenessChecker) Check(request *http.Request) error {
	select {
	case <-l.StopCh:
		return fmt.Errorf("process is shutting down")
	default:
	}
	return nil
}
