package users

type Repo interface {
	PutNewSlackUser(id string, username string, subscribedLines []string) error
	UpdateExistingSlackUser(id string, username string, subscribedLines []string) error
}

type User struct {
	ID              string
	Username        string
	SubscribedLines []string
}
