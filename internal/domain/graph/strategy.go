package graph

type StrategyStep struct {
	WaitBeforeSeconds int
	Action StrategyAction
	Condition StrategyCondition
	WaitBetweenSeconds int
	NumAttempts int
}
