package graph

type FailureStrategy string

const (
	IgnoreFailureStrategy FailureStrategy = "ignore"
	RevertFailureStrategy FailureStrategy = "revert"
)

func ParseFailureStrategy(fs string) FailureStrategy {
	retFs := FailureStrategy(fs)

	switch retFs {
	case IgnoreFailureStrategy,
		RevertFailureStrategy:
		return FailureStrategy(fs)
	default:
		return IgnoreFailureStrategy
	}
}

func (fs FailureStrategy) String() string {
	return string(fs)
}
