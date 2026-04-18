package graph

type StrategyAction string

const (
	SkipStrategyAction StrategyAction = "skip"
	RetryStrategyAction StrategyAction = "retry"
	RevertStrategyAction StrategyAction = "revert"
)

func ParseStrategyAction(sa string) (StrategyAction, bool) {
	retSa := StrategyAction(sa)

	switch retSa {
	case SkipStrategyAction,
		 RetryStrategyAction,
		 RevertStrategyAction:
		return StrategyAction(sa), true
	default:
		return SkipStrategyAction, false
	}
}

func (fs StrategyAction) String() string {
	return string(fs)
}
