package models

type Service interface {
	AddCollaborator(username string) error
	ViewUserProfile(username string) (*Profile, error)
	ShareCode(username, featureName, codeContent string, Token string) error
}

type Profile struct {
	Name  string
	Email string
}
