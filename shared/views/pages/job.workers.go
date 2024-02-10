package pages

type JobWorker struct {
	ID                      string
	Queue                   string
	LastSeenAtColourSuccess bool
	NotSeenSince            string
	Workers                 int
	Version                 string
	JobTypes                []string
}
