package factory

import (
	"errors"

	"github.com/Ammyy9908/gitlib/pkg/github"
	"github.com/Ammyy9908/gitlib/pkg/models"
)

// ServiceType represents the type of the Git service.
type ServiceType string

const (
	// Available service types
	GitHubServiceType    ServiceType = "github"
	GitLabServiceType    ServiceType = "gitlab"
	BitbucketServiceType ServiceType = "bitbucket"
)

// ServiceFactoryOptions contains the required parameters to create a new service.
type ServiceFactoryOptions struct {
	Token string
	Owner string
	Repo  string
}

// NewService creates a new instance of the git service based on the service type.
func NewService(serviceType ServiceType, options ServiceFactoryOptions) (models.Service, error) {
	switch serviceType {
	case GitHubServiceType:
		return github.NewGithubService(options.Token, options.Owner, options.Repo), nil

	default:
		return nil, errors.New("invalid service type")
	}
}
