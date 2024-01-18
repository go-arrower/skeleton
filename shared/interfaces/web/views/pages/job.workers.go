package pages

type JobWorker struct {
	ID               string
	Queue            string
	LastSeenAt       string
	LastSeenAtColour string
	NotSeenSince     string
	Workers          int
}
