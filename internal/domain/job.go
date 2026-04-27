package domain

type VideoJob struct {
	ID         int64
	Filename   string
	RetryCount int
}

func (j VideoJob) CanRetry(max int) bool {
	return j.RetryCount < max
}
