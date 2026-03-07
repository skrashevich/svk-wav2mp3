package batch

// BatchStats contains batch conversion statistics.
type BatchStats struct {
	TotalFiles    int
	Successful    int
	Failed        int
	Skipped       int
	OutputFiles   []string
	ErrorMessages []string
	ElapsedTime   int64 // in seconds
}
