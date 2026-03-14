package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

type GitHubService struct {
	client *github.Client
}

func NewGitHubService(token string) *GitHubService {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return &GitHubService{
		client: github.NewClient(tc),
	}
}

// FetchPipelineLogs is a mocked version. In a real scenario, we would use GitHub Actions API.
func (s *GitHubService) FetchPipelineLogs(ctx context.Context, owner, repo string, runID int64) (string, error) {
	// Mock returning logs
	return "Mock Logs: build failed due to dependency conflict...", nil
}

// FetchCommitDiff gets the diff for a commit
func (s *GitHubService) FetchCommitDiff(ctx context.Context, owner, repo string, sha string) (string, error) {
	commit, _, err := s.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return "", err
	}

	// Create a pseudo-diff from the commit files for the agent to analyze
	var diffBuilder strings.Builder
	for _, file := range commit.Files {
		diffBuilder.WriteString(fmt.Sprintf("File: %s\n", *file.Filename))
		if file.Patch != nil {
			diffBuilder.WriteString(*file.Patch + "\n")
		}
	}
	return diffBuilder.String(), nil
}

func (s *GitHubService) CreatePullRequest(ctx context.Context, owner, repo, baseBranch, headBranch, title, body, patch string) (*github.PullRequest, error) {
	// 1. In a real system, you would first get the base branch SHA
	// 2. Create headBranch referencing that SHA
	// 3. Apply the 'patch' creating a new commit (e.g. using git trees/commits APIs)
	// 4. Create the PR

	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(headBranch),
		Base:                github.String(baseBranch),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := s.client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %v", err)
	}

	return pr, nil
}
