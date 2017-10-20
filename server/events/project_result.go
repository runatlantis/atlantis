package events

type ProjectResult struct {
	Path         string
	Error        error
	Failure      string
	PlanSuccess  *PlanSuccess
	ApplySuccess string
}

func (p ProjectResult) Status() Status {
	if p.Error != nil {
		return Error
	}
	if p.Failure != "" {
		return Failure
	}
	return Success
}
