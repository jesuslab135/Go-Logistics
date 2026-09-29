package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// drainTimeout is how long shutdown waits for a tick in flight before it
// cancels it. A cancelled run is recorded as failed and retried next start.
const drainTimeout = 60 * time.Second

// Start runs r.Tick on the given cron expression until the returned stop
// function is called. stop blocks until the tick in flight has finished, so
// the caller can close the database pool after it returns.
func Start(r *Runner, expression string, logger *slog.Logger) (stop func(), err error) {
	return start(r.Tick, expression, logger, drainTimeout)
}

func start(tick func(context.Context) error, expression string, logger *slog.Logger, drain time.Duration) (func(), error) {
	if logger == nil {
		logger = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())

	// SkipIfStillRunning: a tick that outlasts the interval is not joined by a
	// second one. The advisory lock would refuse it anyway; this saves the
	// connection.
	cronLog := slogCronLogger{logger}
	c := cron.New(cron.WithChain(
		cron.Recover(cronLog),
		cron.SkipIfStillRunning(cronLog),
	))
	_, err := c.AddFunc(expression, func() {
		if err := tick(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				// Shutdown cancelled the tick in flight; that is expected,
				// not a failure, and must not be logged as one.
				logger.Info("scheduler: tick cancelled at shutdown")
				return
			}
			logger.Error("scheduler: tick failed", "error", err)
		}
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("scheduler: REPORTS_TICK %q is not a cron expression: %w", expression, err)
	}
	c.Start()
	logger.Info("scheduler: started", "tick", expression)

	return func() {
		done := c.Stop().Done()
		select {
		case <-done:
		case <-time.After(drain):
			logger.Warn("scheduler: tick still running at shutdown, cancelling it")
			cancel()
			<-done
		}
		cancel()
		logger.Info("scheduler: stopped")
	}, nil
}

// slogCronLogger adapts a *slog.Logger to cron.Logger, so a panic recovered
// from a tick (cron.Recover) and a tick skipped because the previous one is
// still running (cron.SkipIfStillRunning) are logged instead of discarded.
type slogCronLogger struct {
	logger *slog.Logger
}

func (l slogCronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l slogCronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, append([]interface{}{"error", err}, keysAndValues...)...)
}
