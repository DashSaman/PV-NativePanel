package customer

// Service-layer aliases keep request terminology concise while preserving the
// database's existing canonical start_policy values.
const (
	StartOnFirstConnection StartPolicy = StartOnFirstSuccessfulConnection
	StartFixedTimestamp    StartPolicy = StartAtFixedTimestamp
)
