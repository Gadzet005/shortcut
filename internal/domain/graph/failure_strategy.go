package graph

type FailureStrategy string

const (
	AbsentFailureStrategy FailureStrategy = "absent"
	IgnoreFailureStrategy FailureStrategy = "ignore"
	RevertFailureStrategy FailureStrategy = "revert"
	SaveFailureStrategy   FailureStrategy = "save"
	CustomFailureStrategy FailureStrategy = "custom"
)

func ParseFailureStrategy(fs string) (FailureStrategy, bool) {
	retFs := FailureStrategy(fs)

	switch retFs {
	case IgnoreFailureStrategy,
		RevertFailureStrategy,
		SaveFailureStrategy,
		CustomFailureStrategy:
		return FailureStrategy(fs), true
	default:
		return IgnoreFailureStrategy, false
	}
}

func (fs FailureStrategy) String() string {
	return string(fs)
}
