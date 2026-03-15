package optional

func New[Data any](data Data) T[Data] {
	return T[Data]{
		data:    data,
		present: true,
	}
}

func Empty[Data any]() T[Data] {
	return T[Data]{}
}

func FromPointer[Data any](data *Data) T[Data] {
	if data == nil {
		return Empty[Data]()
	}
	return New(*data)
}

type T[Data any] struct {
	data    Data
	present bool
}

func (t T[Data]) IsPresent() bool {
	return t.present
}

func (t T[Data]) Value() (Data, bool) {
	return t.data, t.present
}

func (t T[Data]) Or(defaultValue Data) Data {
	if t.present {
		return t.data
	}
	return defaultValue
}
