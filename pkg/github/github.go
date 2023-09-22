package github

import (
	"context"
	"fmt"

	"github.com/Ammyy9908/gitlib/pkg/models"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type GithubService struct {
	client  *github.Client
	owner   string
	repo    string
	context context.Context
}

func NewGithubService(token, owner, repo string) *GithubService {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := github.NewClient(oauthClient)
	return &GithubService{
		client:  client,
		owner:   owner,
		repo:    repo,
		context: context.Background(),
	}

}

func (g *GithubService) AddCollaborator(username string) error {
	_, _, err := g.client.Repositories.AddCollaborator(context.Background(), g.owner, g.repo, username, nil)
	return err
}

func (g *GithubService) ViewUserProfile(username string) (*models.Profile, error) {
	user, _, err := g.client.Users.Get(context.Background(), username)
	return &models.Profile{Name: user.GetName(), Email: user.GetEmail()}, err
}

func (g *GithubService) ShareCode(username, featureName, codeContent string) error {
	branchName := fmt.Sprintf("%s-%s", username, featureName)

	// 1. Create a new branch
	ref, _, err := g.client.Git.GetRef(g.context, g.owner, g.repo, "refs/heads/main")
	if err != nil {
		return err
	}

	newRef := &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: &github.GitObject{SHA: ref.Object.SHA},
	}
	_, _, err = g.client.Git.CreateRef(g.context, g.owner, g.repo, newRef)
	if err != nil {
		return err
	}

	// 2. Commit the code to the new branch
	filePath := username + "/code.txt"
	fileContent := []byte(codeContent)

	// Create a new tree with the file
	entries := []*github.TreeEntry{
		{
			Path:    github.String(filePath),
			Type:    github.String("blob"),
			Content: github.String(string(fileContent)),
			Mode:    github.String("100644"), // This indicates it's a file
		},
	}

	tree, _, err := g.client.Git.CreateTree(g.context, g.owner, g.repo, *ref.Object.SHA, entries)
	if err != nil {
		return err
	}

	newListOptions := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	// Commit the tree
	parentCommit, _, err := g.client.Repositories.GetCommit(g.context, g.owner, g.repo, *ref.Object.SHA, newListOptions)
	if err != nil {
		return err
	}

	parentCommitCore := parentCommit.GetCommit()
	commitMessage := fmt.Sprintf("Added code by %s", username)
	newCommit := &github.Commit{
		Message: github.String(commitMessage),
		Tree:    tree,
		Parents: []*github.Commit{parentCommitCore},
	}

	commit, _, err := g.client.Git.CreateCommit(g.context, g.owner, g.repo, newCommit)
	if err != nil {
		return err
	}

	// Push the commit to the new branch
	_, _, err = g.client.Git.UpdateRef(g.context, g.owner, g.repo, &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: &github.GitObject{SHA: commit.SHA},
	}, false)
	if err != nil {
		return err
	}

	// 3. Create a pull request
	prTitle := "Code shared by " + username
	newPR := &github.NewPullRequest{
		Title:               github.String(prTitle),
		Head:                github.String(branchName),
		Base:                github.String("main"),
		Body:                github.String(fmt.Sprintf("Code shared by %s for feature %s", username, featureName)),
		MaintainerCanModify: github.Bool(true),
	}

	_, _, err = g.client.PullRequests.Create(g.context, g.owner, g.repo, newPR)
	if err != nil {
		return err
	}

	return nil
}
