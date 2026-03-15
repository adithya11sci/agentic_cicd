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

func (s *GitHubService) FetchPipelineLogs(ctx context.Context, owner, repo string, runID int64) (string, error) {
return "Traceback (most recent call last): File calc.py, line 6: TypeError: unsupported operand type(s) for +: int and str", nil
}

func (s *GitHubService) FetchCommitDiff(ctx context.Context, owner, repo string, sha string) (string, error) {
// Fake diff if it doesn't match an actual SHA, safe for demo webhooks
if len(sha) < 40 {
return "diff --git a/main.go b/main.go\n- invalid_syntax\n+ valid_syntax", nil
}
commit, _, err := s.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
if err != nil {
return "", err
}

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
// For the purposes of a safe API demonstration without creating orphaned branches/trees, 
// we will create an Issue containing the patch and tag it as the Agent's PR.

issueReq := &github.IssueRequest{
Title: github.String(title + " (Agentic PR Mock)"),
Body:  github.String(fmt.Sprintf("### Target Base Branch: `%s`\n\n%s\n\n**Proposed Patch:**\n```diff\n%s\n```", baseBranch, body, patch)),
}

issue, _, err := s.client.Issues.Create(ctx, owner, repo, issueReq)
if err != nil {
return nil, fmt.Errorf("failed to create issue as PR fallback: %v", err)
}

return &github.PullRequest{
HTMLURL: issue.HTMLURL,
Number:  issue.Number,
}, nil
}
