package strategyworker

import (
	"context"
	"sync"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/Gadzet005/shortcut/internal/domain/trace"
	"go.uber.org/zap"
)

type GraphResolver interface {
	GetFailureSteps(namespaceID graph.NamespaceID, graphID graph.ID) ([]graph.FailureStep, error)
}

type Config struct {
	Interval          time.Duration
	BatchSize         int
	VisibilityTimeout time.Duration
}

type Worker struct {
	failureRepo failure.Repo
	traceRepo   trace.Repo
	recovery    failure.Recovery
	resolver    GraphResolver
	logger      *zap.Logger
	cfg         Config
}

func New(
	failureRepo failure.Repo,
	traceRepo trace.Repo,
	recovery failure.Recovery,
	resolver GraphResolver,
	logger *zap.Logger,
	cfg Config,
) *Worker {
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 32
	}
	if cfg.VisibilityTimeout <= 0 {
		cfg.VisibilityTimeout = 30 * time.Second
	}
	return &Worker{
		failureRepo: failureRepo,
		traceRepo:   traceRepo,
		recovery:    recovery,
		resolver:    resolver,
		logger:      logger.Named("strategy-worker"),
		cfg:         cfg,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("strategy worker started",
		zap.Duration("interval", w.cfg.Interval),
		zap.Int("batch_size", w.cfg.BatchSize),
		zap.Duration("visibility_timeout", w.cfg.VisibilityTimeout),
	)

	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("strategy worker stopping")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	batch, err := w.failureRepo.ClaimReadyBatch(ctx, w.cfg.BatchSize, w.cfg.VisibilityTimeout)
	if err != nil {
		w.logger.Error("claim ready batch failed", zap.Error(err))
		return
	}
	if len(batch) == 0 {
		return
	}

	w.logger.Debug("claimed batch", zap.Int("size", len(batch)))

	var wg sync.WaitGroup
	for _, f := range batch {
		wg.Add(1)
		go func(f failure.Failure) {
			defer wg.Done()
			w.processOne(ctx, f)
		}(f)
	}
	wg.Wait()
}

func (w *Worker) processOne(ctx context.Context, f failure.Failure) {
	steps, err := w.resolver.GetFailureSteps(graph.NamespaceID(f.NamespaceID), graph.ID(f.GraphID))
	if err != nil {
		w.logger.Error("resolve failure steps failed",
			zap.String("request_id", f.RequestID),
			zap.String("namespace_id", f.NamespaceID),
			zap.String("graph_id", f.GraphID),
			zap.Error(err))
		return
	}

	stepIdx := int(f.NumRetry)
	if stepIdx >= len(steps) {
		w.markDone(ctx, f)
		return
	}
	step := steps[stepIdx]

	w.logger.Info("processing failure",
		zap.String("request_id", f.RequestID),
		zap.String("namespace_id", f.NamespaceID),
		zap.String("graph_id", f.GraphID),
		zap.Int("step_idx", stepIdx),
		zap.String("step_action", string(step.Action)),
		zap.String("step_condition", string(step.Condition)),
		zap.Int64("num_retry", f.NumRetry),
	)

	if !w.conditionMet(step.Condition, f) {
		w.logger.Info("step condition not met, skipping",
			zap.String("request_id", f.RequestID),
			zap.String("step_condition", string(step.Condition)),
		)
		w.advance(ctx, f, len(steps), 0)
		return
	}

	actionSucceeded, terminal := w.runAction(ctx, step, f)

	if terminal {
		w.finalize(ctx, f, actionSucceeded)
		return
	}

	w.advance(ctx, f, len(steps), step.WaitBefore)
}

func (w *Worker) conditionMet(cond graph.StrategyCondition, f failure.Failure) bool {
	switch cond {
	case graph.NoStrategyCondition:
		return true
	case graph.LastActionSuccessfulStrategyCondition:
		return f.NumRetry > 0 && f.Status == failure.StatusProcessing
	case graph.LastActionFailedStrategyCondition:
		return f.NumRetry > 0 && f.Status == failure.StatusFailed
	default:
		return true
	}
}

func (w *Worker) runAction(ctx context.Context, step graph.FailureStep, f failure.Failure) (bool, bool) {
	switch step.Action {
	case graph.SkipStrategyAction:
		w.logger.Info("skip action applied", zap.String("request_id", f.RequestID))
		return true, false

	case graph.RevertStrategyAction:
		ok, err := w.recovery.Revert(ctx, f.NamespaceID, f.GraphID, f.RequestID, f.VisitedNodes())
		if err != nil {
			w.logger.Error("revert action failed", zap.String("request_id", f.RequestID), zap.Error(err))
			return false, false
		}
		w.logger.Info("revert action finished", zap.String("request_id", f.RequestID), zap.Bool("ok", ok))
		return ok, true

	case graph.FinishStrategyAction:
		ok, err := w.recovery.Finish(ctx, f.NamespaceID, f.GraphID, f.RequestID, f.VisitedNodes(), f.Method, f.Path, f.RequestBody)
		if err != nil {
			w.logger.Error("finish action failed", zap.String("request_id", f.RequestID), zap.Error(err))
			return false, false
		}
		w.logger.Info("finish action finished", zap.String("request_id", f.RequestID), zap.Bool("ok", ok))
		return ok, ok

	case graph.RetryStrategyAction:
		ok := w.runRetry(ctx, step, f)
		w.logger.Info("retry action finished", zap.String("request_id", f.RequestID), zap.Bool("ok", ok))
		return ok, false

	default:
		w.logger.Warn("unknown strategy action", zap.String("action", string(step.Action)))
		return false, false
	}
}

func (w *Worker) runRetry(ctx context.Context, step graph.FailureStep, f failure.Failure) bool {
	attempts := step.NumAttempts
	if attempts <= 0 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		if i > 0 && step.WaitBetweenRetries > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(step.WaitBetweenRetries):
			}
		}
		ok, err := w.recovery.Retry(ctx, f.NamespaceID, f.GraphID, f.Method, f.Path, f.RequestBody)
		if err != nil {
			w.logger.Warn("retry attempt failed",
				zap.String("request_id", f.RequestID),
				zap.Int("attempt", i+1),
				zap.Error(err))
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func (w *Worker) advance(ctx context.Context, f failure.Failure, totalSteps int, waitBefore time.Duration) {
	nextRetry := f.NumRetry + 1
	status := failure.StatusProcessing
	var nextAt time.Time
	if int(nextRetry) >= totalSteps {
		status = failure.StatusDone
		nextAt = time.Now()
	} else {
		nextAt = time.Now().Add(waitBefore)
	}

	if err := w.failureRepo.UpdateProgress(ctx, f.RequestID, nextRetry, nextAt, status); err != nil {
		w.logger.Error("update progress failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return
	}

	if status == failure.StatusDone {
		w.logger.Info("failure done",
			zap.String("request_id", f.RequestID),
			zap.Int64("num_retry", nextRetry),
			zap.String("reason", "all steps consumed"),
		)
		w.deleteCompleted(ctx, f)
	}
}

func (w *Worker) finalize(ctx context.Context, f failure.Failure, success bool) {
	if !success {
		nextAt := time.Now().Add(w.cfg.VisibilityTimeout)
		if err := w.failureRepo.UpdateProgress(ctx, f.RequestID, f.NumRetry+1, nextAt, failure.StatusFailed); err != nil {
			w.logger.Error("finalize update failed", zap.String("request_id", f.RequestID), zap.Error(err))
		}
		return
	}
	if err := w.failureRepo.UpdateProgress(ctx, f.RequestID, f.NumRetry+1, time.Now(), failure.StatusDone); err != nil {
		w.logger.Error("finalize update failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return
	}
	w.logger.Info("failure done",
		zap.String("request_id", f.RequestID),
		zap.Int64("num_retry", f.NumRetry+1),
		zap.String("reason", "terminal action succeeded"),
	)
	w.deleteCompleted(ctx, f)
}

func (w *Worker) deleteCompleted(ctx context.Context, f failure.Failure) {
	if err := w.failureRepo.Delete(ctx, f.RequestID); err != nil {
		w.logger.Error("delete failure record failed", zap.String("request_id", f.RequestID), zap.Error(err))
	}
	if w.traceRepo == nil {
		return
	}
	if deleter, ok := w.traceRepo.(traceDeleter); ok {
		if err := deleter.DeleteByRequestID(ctx, trace.RequestID(f.RequestID)); err != nil {
			w.logger.Warn("delete trace failed", zap.String("request_id", f.RequestID), zap.Error(err))
		}
	}
}

func (w *Worker) markDone(ctx context.Context, f failure.Failure) {
	if err := w.failureRepo.UpdateProgress(ctx, f.RequestID, f.NumRetry, time.Now(), failure.StatusDone); err != nil {
		w.logger.Error("mark done failed", zap.String("request_id", f.RequestID), zap.Error(err))
		return
	}
	w.logger.Info("failure done",
		zap.String("request_id", f.RequestID),
		zap.Int64("num_retry", f.NumRetry),
		zap.String("reason", "no steps remaining"),
	)
	w.deleteCompleted(ctx, f)
}

type traceDeleter interface {
	DeleteByRequestID(ctx context.Context, requestID trace.RequestID) error
}
