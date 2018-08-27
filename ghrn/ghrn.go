package ghrn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Config describes configuration for BuildReleaseNotes.
type Config struct {
	// Org is the name of the GitHub organization. Required.
	Org string
	// Repo is the name of the GitHub repository. Required.
	Repo string

	// GitHubToken is a GitHub API access token.
	GitHubToken string

	// StopAt is the number of the Pull Request to stop at.
	// Useful for building the notes of PRs since the last release, for example.
	StopAt int
	// IncludeCommits will include commmits messages for each PR.
	IncludeCommits bool
	// SinceLatestRelease will only include PRs and commits merged since the latest release tag.
	SinceLatestRelease bool
	// IncludeAuthor will prefix the message with an author of the PR
	IncludeAuthor bool
}

// BuildReleaseNotes lists GitHub Pull Requests and writes formatted release notes
// to the given writer.
func BuildReleaseNotes(ctx context.Context, w io.Writer, conf Config) error {

	if conf.Org == "" {
		return fmt.Errorf("Config.Org is required")
	}
	if conf.Repo == "" {
		return fmt.Errorf("Config.Repo is required")
	}

	var httpClient *http.Client
	if conf.GitHubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: conf.GitHubToken},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}
	cl := github.NewClient(httpClient)

	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "closed",
	}

	repo, _, err := cl.Repositories.Get(ctx, conf.Org, conf.Repo)
	if err != nil {
		return fmt.Errorf("get repository: %+v", err)
	}
	var commitsNotMerged []string = nil
	// Iterate over all PRs
	for {
		prs, resp, err := cl.PullRequests.List(ctx, conf.Org, conf.Repo, opt)
		if err != nil {
			return fmt.Errorf("listing PRs: %s", err)
		}

		// Iterate over PRs in this page.
		for _, pr := range prs {
			if *pr.Number == conf.StopAt {
				return nil
			}
			if pr.MergedAt == nil {
				continue
			}

			commits, err := commitsAll(ctx, cl, conf.Org, conf.Repo, pr.GetNumber())
			if err != nil {
				return fmt.Errorf("listing PR commits: %s", err)
			}

			if conf.SinceLatestRelease {
				if pr.GetBase().GetRef() != repo.GetDefaultBranch() {
					// Skip when PR base branch isn't a default branch
					continue
				}

				if commitsNotMerged == nil {
					commitsNotMerged, err = newCommits(ctx, cl, conf.Org, conf.Repo)
					if err != nil {
						return fmt.Errorf("listing new commits: %+v", err)
					}
				}
				if !any(commitHashes(commits), commitsNotMerged) {
					// Stop when a PR doesn't contain any commits from since the latest release.
					return nil
				}
			}

			if conf.IncludeAuthor {
				fmt.Fprintf(w, "- PR #%d - @%s - %s\n", pr.GetNumber(), *pr.GetUser().Login, pr.GetTitle())
			} else {
				fmt.Fprintf(w, "- PR #%d %s\n", pr.GetNumber(), pr.GetTitle())
			}

			if conf.IncludeCommits {
				// Iterate over all commits in this PR.
				for _, commit := range commits {
					sha := *commit.SHA
					msg := *commit.Commit.Message

					// Strip multiple lines (i.e. only take first line)
					if i := strings.Index(msg, "\n"); i != -1 {
						msg = msg[:i]
					}
					// Trim long lines
					if len(msg) > 90 {
						msg = msg[:90] + "..."
					}
					msg = strings.TrimSpace(msg)

					fmt.Fprintf(w, "    - %s %s\n", sha, msg)
				}
				fmt.Fprintln(w)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

func contains(a []string, e string) bool {
	for _, v := range a {
		if e == v {
			return true
		}
	}
	return false
}

func any(a []string, b []string) bool {
	for _, c := range a {
		if contains(b, c) {
			return true
		}
	}
	return false
}

func commitsAll(ctx context.Context, cl *github.Client, owner string, repo string, num int) ([]github.RepositoryCommit, error) {
	var list []github.RepositoryCommit
	commitOpt := &github.ListOptions{PerPage: 100}
	for {
		commits, resp, err := cl.PullRequests.ListCommits(ctx, owner, repo, num, commitOpt)
		if err != nil {
			return nil, fmt.Errorf("listing PR commits: %s", err)
		}

		for _, commit := range commits {
			list = append(list, *commit)
		}

		if resp.NextPage == 0 {
			break
		}
		commitOpt.Page = resp.NextPage
	}
	return list, nil
}

func commitHashes(commits []github.RepositoryCommit) []string {
	var newCommits []string
	for _, commit := range commits {
		newCommits = append(newCommits, commit.GetCommit().GetTree().GetSHA())
	}
	return newCommits
}

func newCommits(ctx context.Context, cl *github.Client, owner string, repo string) ([]string, error) {
	repository, _, err := cl.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get repository: %+v", err)
	}

	rls, _, err := cl.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get latest release: %+v", err)
	}

	comp, _, err := cl.Repositories.CompareCommits(ctx, owner, repo, rls.GetTagName(), repository.GetDefaultBranch())
	if err != nil {
		return nil, fmt.Errorf("compare commitse: %s..%s %+v", rls.GetTagName(), repository.GetDefaultBranch(), err)
	}

	return commitHashes(comp.Commits), nil
}
