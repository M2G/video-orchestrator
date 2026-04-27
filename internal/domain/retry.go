package domain

import "time"

func NextDelay(retry int) int {

	base := 5
	max := 300

	delay := base * (1 << retry)

	if delay > max {
		delay = max
	}

	jitter := time.Now().UnixNano() % 5

	return delay + int(jitter)
}
