package ibdf

type MissingStateError struct {
}

func (e *MissingStateError) Error() string {
	return "missing state in packet file"
}
