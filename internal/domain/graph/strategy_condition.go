package graph

type StrategyCondition string

const (
	AlwaysStrategyCondition StrategyCondition = "always"
	LastActionSuccessfulStrategyCondition StrategyCondition = "last_action_successful"
)

func ParseStrategyCondition(sc string) (StrategyCondition, bool) {
	retSc := StrategyCondition(sc)

	switch retSc {
	case AlwaysStrategyCondition,
		 LastActionSuccessfulStrategyCondition:
		return StrategyCondition(sc), true
	default:
		return AlwaysStrategyCondition, false
	}
}

func (sc StrategyCondition) String() string {
	return string(sc)
}