package payment

func BackoffMs(attempt int) int {
	ms := 100 << attempt
	if ms > 2000 {
		return 2000
	}
	return ms
}
