package strategyworker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeRepo struct {
	mu         sync.Mutex
	claim      []failure.Failure
	progress   []progressCall
	deleted    []string
	getRequest failure.Failure
	saveCalls  int
}

type progressCall struct {
	requestID      string
	numRetry       int64
	readyToRetryAt time.Time
	status         failure.Status
}

func (r *fakeRepo) Save(_ context.Context, _ failure.Failure) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveCalls++
	return nil
}

func (r *fakeRepo) GetByRequestID(_ context.Context, _ string) (failure.Failure, error) {
	return r.getRequest, nil
}

func (r *fakeRepo) Delete(_ context.Context, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleted = append(r.deleted, requestID)
	return nil
}

func (r *fakeRepo) ListByGraph(_ context.Context, _, _ string) ([]failure.Failure, error) {
	return nil, nil
}

func (r *fakeRepo) ClaimReadyBatch(_ context.Context, _ int, _ time.Duration) ([]failure.Failure, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.claim
	r.claim = nil
	return out, nil
}

func (r *fakeRepo) UpdateProgress(_ context.Context, requestID string, numRetry int64, readyAt time.Time, status failure.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.progress = append(r.progress, progressCall{requestID, numRetry, readyAt, status})
	return nil
}

type fakeRecovery struct {
	mu          sync.Mutex
	revertOK    bool
	revertErr   error
	retryOK     bool
	finishOK    bool
	revertCalls int
	retryCalls  int
	finishCalls int
}

func (r *fakeRecovery) Revert(_ context.Context, _, _, _ string, _ []string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revertCalls++
	return r.revertOK, r.revertErr
}

func (r *fakeRecovery) Retry(_ context.Context, _, _, _, _ string, _ []byte) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.retryCalls++
	return r.retryOK, nil
}

func (r *fakeRecovery) Finish(_ context.Context, _, _, _ string, _ []string, _, _ string, _ []byte) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finishCalls++
	return r.finishOK, nil
}

type fakeResolver struct {
	steps []graph.FailureStep
	err   error
}

func (f *fakeResolver) GetFailureSteps(_ graph.NamespaceID, _ graph.ID) ([]graph.FailureStep, error) {
	return f.steps, f.err
}

func TestWorker_RetryActionAdvancesStep(t *testing.T) {
	steps := []graph.FailureStep{
		{Action: graph.RetryStrategyAction, NumAttempts: 1, WaitBefore: 0},
		{Action: graph.SkipStrategyAction, WaitBefore: 0},
	}
	repo := &fakeRepo{}
	rec := &fakeRecovery{retryOK: false}
	resolver := &fakeResolver{steps: steps}

	w := New(repo, nil, rec, resolver, zap.NewNop(), Config{})
	w.processOne(context.Background(), failure.Failure{
		RequestID: "req1", NumRetry: 0, NamespaceID: "ns", GraphID: "g",
	})

	assert.Equal(t, 1, rec.retryCalls)
	require.Len(t, repo.progress, 1)
	assert.EqualValues(t, 1, repo.progress[0].numRetry)
	assert.Equal(t, failure.StatusProcessing, repo.progress[0].status)
}

func TestWorker_RetrySuccess_DoesNotShortCircuit(t *testing.T) {
	steps := []graph.FailureStep{
		{Action: graph.RetryStrategyAction, NumAttempts: 1, WaitBefore: 0},
	}
	repo := &fakeRepo{}
	rec := &fakeRecovery{retryOK: true}
	resolver := &fakeResolver{steps: steps}

	w := New(repo, nil, rec, resolver, zap.NewNop(), Config{})
	w.processOne(context.Background(), failure.Failure{
		RequestID: "req1", NamespaceID: "ns", GraphID: "g",
	})

	assert.Equal(t, 1, rec.retryCalls)
	require.Len(t, repo.progress, 1)
	assert.Equal(t, failure.StatusDone, repo.progress[0].status)
	assert.Equal(t, []string{"req1"}, repo.deleted)
}

func TestWorker_RevertAction_TerminatesOnSuccess(t *testing.T) {
	steps := []graph.FailureStep{
		{Action: graph.RevertStrategyAction},
	}
	repo := &fakeRepo{}
	rec := &fakeRecovery{revertOK: true}
	resolver := &fakeResolver{steps: steps}

	w := New(repo, nil, rec, resolver, zap.NewNop(), Config{})
	w.processOne(context.Background(), failure.Failure{
		RequestID: "req1", NamespaceID: "ns", GraphID: "g",
	})

	assert.Equal(t, 1, rec.revertCalls)
	require.Len(t, repo.progress, 1)
	assert.Equal(t, failure.StatusDone, repo.progress[0].status)
	assert.Equal(t, []string{"req1"}, repo.deleted)
}

func TestWorker_FinishAction_DoesNotTerminateOnFailure(t *testing.T) {
	steps := []graph.FailureStep{
		{Action: graph.FinishStrategyAction},
		{Action: graph.SkipStrategyAction},
	}
	repo := &fakeRepo{}
	rec := &fakeRecovery{finishOK: false}
	resolver := &fakeResolver{steps: steps}

	w := New(repo, nil, rec, resolver, zap.NewNop(), Config{})
	w.processOne(context.Background(), failure.Failure{
		RequestID: "req1", NamespaceID: "ns", GraphID: "g",
	})

	assert.Equal(t, 1, rec.finishCalls)
	require.Len(t, repo.progress, 1)
	assert.Equal(t, failure.StatusProcessing, repo.progress[0].status)
	assert.Empty(t, repo.deleted)
}

func TestWorker_NoMoreSteps_MarksDone(t *testing.T) {
	repo := &fakeRepo{}
	resolver := &fakeResolver{steps: []graph.FailureStep{{Action: graph.SkipStrategyAction}}}

	w := New(repo, nil, &fakeRecovery{}, resolver, zap.NewNop(), Config{})
	w.processOne(context.Background(), failure.Failure{
		RequestID: "req1", NumRetry: 5, NamespaceID: "ns", GraphID: "g",
	})

	require.Len(t, repo.progress, 1)
	assert.Equal(t, failure.StatusDone, repo.progress[0].status)
	assert.Equal(t, []string{"req1"}, repo.deleted)
}

func TestWorker_TickProcessesBatchInParallel(t *testing.T) {
	repo := &fakeRepo{
		claim: []failure.Failure{
			{RequestID: "a", NamespaceID: "ns", GraphID: "g"},
			{RequestID: "b", NamespaceID: "ns", GraphID: "g"},
			{RequestID: "c", NamespaceID: "ns", GraphID: "g"},
		},
	}
	rec := &fakeRecovery{revertOK: true}
	resolver := &fakeResolver{steps: []graph.FailureStep{{Action: graph.RevertStrategyAction}}}

	w := New(repo, nil, rec, resolver, zap.NewNop(), Config{})
	w.tick(context.Background())

	assert.Equal(t, 3, rec.revertCalls)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, repo.deleted)
}

func TestWorker_Defaults(t *testing.T) {
	w := New(&fakeRepo{}, nil, &fakeRecovery{}, &fakeResolver{}, zap.NewNop(), Config{})
	assert.Equal(t, 5*time.Second, w.cfg.Interval)
	assert.Equal(t, 32, w.cfg.BatchSize)
	assert.Equal(t, 30*time.Second, w.cfg.VisibilityTimeout)
}
