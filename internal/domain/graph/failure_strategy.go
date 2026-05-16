package graph

import "time"

type FailureStrategy string

const (
	AbsentFailureStrategy FailureStrategy = "absent"
	IgnoreFailureStrategy FailureStrategy = "ignore"
	RevertFailureStrategy FailureStrategy = "revert"
	SaveFailureStrategy   FailureStrategy = "save"
	FinishFailureStrategy FailureStrategy = "finish"
	CustomFailureStrategy FailureStrategy = "custom"
)

func ParseFailureStrategy(fs string) (FailureStrategy, bool) {
	retFs := FailureStrategy(fs)

	switch retFs {
	case IgnoreFailureStrategy,
		RevertFailureStrategy,
		SaveFailureStrategy,
		FinishFailureStrategy,
		CustomFailureStrategy,
		AbsentFailureStrategy:
		return retFs, true
	default:
		return AbsentFailureStrategy, false
	}
}

func (fs FailureStrategy) String() string {
	return string(fs)
}

type StrategyAction string

const (
	SkipStrategyAction   StrategyAction = "skip"
	RetryStrategyAction  StrategyAction = "retry"
	RevertStrategyAction StrategyAction = "revert"
	FinishStrategyAction StrategyAction = "finish"
)

func ParseStrategyAction(s string) (StrategyAction, bool) {
	a := StrategyAction(s)
	switch a {
	case SkipStrategyAction,
		RetryStrategyAction,
		RevertStrategyAction,
		FinishStrategyAction:
		return a, true
	default:
		return "", false
	}
}

type StrategyCondition string

const (
	NoStrategyCondition                   StrategyCondition = ""
	LastActionSuccessfulStrategyCondition StrategyCondition = "last_action_successful"
	LastActionFailedStrategyCondition     StrategyCondition = "last_action_failed"
)

type FailureStep struct {
	Action             StrategyAction
	Condition          StrategyCondition
	WaitBefore         time.Duration
	WaitBetweenRetries time.Duration
	NumAttempts        int
}
