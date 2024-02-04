package pages

type JobWorker struct {
	ID               string
	Queue            string
	LastSeenAtColour string
	NotSeenSince     string
	Workers          int
	Version          string
	JobTypes         []string
}
