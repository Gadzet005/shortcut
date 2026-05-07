package strategy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Gadzet005/shortcut/internal/domain/failure"
	"github.com/Gadzet005/shortcut/internal/domain/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeRepo struct {
	saved      []failure.Failure
	saveErr    error
	updateErr  error
	deleteErr  error
	progress   []progressCall
	getResult  failure.Failure
	getErr     error
	listResult []failure.Failure
	listErr    error
	claimResult []failure.Failure
	claimErr   error
}

type progressCall struct {
	requestID      string
	numRetry       int64
	readyToRetryAt time.Time
	status         failure.Status
}

func (r *fakeRepo) Save(_ context.Context, f failure.Failure) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = append(r.saved, f)
	return nil
}

func (r *fakeRepo) GetByRequestID(_ context.Context, _ string) (failure.Failure, error) {
	return r.getResult, r.getErr
}

func (r *fakeRepo) Delete(_ context.Context, _ string) error {
	return r.deleteErr
}

func (r *fakeRepo) ListByGraph(_ context.Context, _, _ string) ([]failure.Failure, error) {
	return r.listResult, r.listErr
}

func (r *fakeRepo) ClaimReadyBatch(_ context.Context, _ int, _ time.Duration) ([]failure.Failure, error) {
	return r.claimResult, r.claimErr
}

func (r *fakeRepo) UpdateProgress(_ context.Context, requestID string, numRetry int64, readyAt time.Time, status failure.Status) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.progress = append(r.progress, progressCall{requestID, numRetry, readyAt, status})
	return nil
}

type fakeRecovery struct {
	revertOK   bool
	revertErr  error
	retryOK    bool
	retryErr   error
	finishOK   bool
	finishErr  error
	revertCalls []string
	retryCalls  int
	finishCalls []string
}

func (r *fakeRecovery) Revert(_ context.Context, _, _, requestID string, _ []string) (bool, error) {
	r.revertCalls = append(r.revertCalls, requestID)
	return r.revertOK, r.revertErr
}

func (r *fakeRecovery) Retry(_ context.Context, _, _, _, _ string, _ []byte) (bool, error) {
	r.retryCalls++
	return r.retryOK, r.retryErr
}

func (r *fakeRecovery) Finish(_ context.Context, _, _, requestID string, _ []string, _, _ string, _ []byte) (bool, error) {
	r.finishCalls = append(r.finishCalls, requestID)
	return r.finishOK, r.finishErr
}

func newFactory(repo failure.Repo, recovery failure.Recovery) Factory {
	return NewFactory(repo, recovery, zap.NewNop())
}

func TestFactory_DispatchesByStrategy(t *testing.T) {
	repo := &fakeRepo{}
	rec := &fakeRecovery{revertOK: true, finishOK: true, retryOK: true}
	f := newFactory(repo, rec)

	cases := []struct {
		name     string
		strategy graph.FailureStrategy
		want     interface{}
	}{
		{"ignore", graph.IgnoreFailureStrategy, ignoreHandler{}},
		{"absent", graph.AbsentFailureStrategy, absentHandler{}},
		{"revert", graph.RevertFailureStrategy, revertHandler{}},
		{"save", graph.SaveFailureStrategy, saveHandler{}},
		{"finish", graph.FinishFailureStrategy, finishHandler{}},
		{"custom", graph.CustomFailureStrategy, customHandler{}},
		{"unknown", "unknown", absentHandler{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := f.New(tc.strategy, nil)
			assert.IsType(t, tc.want, h)
		})
	}
}

func TestIgnoreHandler_NoOp(t *testing.T) {
	err := ignoreHandler{}.Handle(context.Background(), failure.Failure{})
	assert.NoError(t, err)
}

func TestAbsentHandler_NoOp(t *testing.T) {
	err := absentHandler{}.Handle(context.Background(), failure.Failure{})
	assert.NoError(t, err)
}

func TestRevertHandler_CallsRecovery(t *testing.T) {
	rec := &fakeRecovery{revertOK: true}
	h := revertHandler{recovery: rec, logger: zap.NewNop()}

	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"req1"}, rec.revertCalls)
}

func TestRevertHandler_PropagatesError(t *testing.T) {
	rec := &fakeRecovery{revertErr: errors.New("boom")}
	h := revertHandler{recovery: rec, logger: zap.NewNop()}

	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})
	assert.Error(t, err)
}

func TestSaveHandler_PersistsAsFailed(t *testing.T) {
	repo := &fakeRepo{}
	h := saveHandler{repo: repo, logger: zap.NewNop()}

	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, failure.StatusFailed, repo.saved[0].Status)
	assert.Equal(t, "req1", repo.saved[0].RequestID)
}

func TestFinishHandler_CallsRecovery(t *testing.T) {
	rec := &fakeRecovery{finishOK: true}
	h := finishHandler{recovery: rec, logger: zap.NewNop()}

	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"req1"}, rec.finishCalls)
}

func TestCustomHandler_PersistsWithFirstStepDelay(t *testing.T) {
	repo := &fakeRepo{}
	steps := []graph.FailureStep{
		{Action: graph.RetryStrategyAction, WaitBefore: 30 * time.Second, NumAttempts: 3},
	}
	h := customHandler{repo: repo, steps: steps, logger: zap.NewNop()}

	before := time.Now()
	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})
	after := time.Now()

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	saved := repo.saved[0]
	assert.Equal(t, failure.StatusPending, saved.Status)
	assert.EqualValues(t, 0, saved.NumRetry)
	assert.True(t, !saved.ReadyToRetryAt.Before(before.Add(30*time.Second)))
	assert.True(t, !saved.ReadyToRetryAt.After(after.Add(30*time.Second)))
}

func TestCustomHandler_NoStepsUsesZeroDelay(t *testing.T) {
	repo := &fakeRepo{}
	h := customHandler{repo: repo, steps: nil, logger: zap.NewNop()}

	before := time.Now()
	err := h.Handle(context.Background(), failure.Failure{RequestID: "req1"})

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.True(t, !repo.saved[0].ReadyToRetryAt.Before(before))
}
