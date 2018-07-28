package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/buchanae/github-release-notes/ghrn"
)

func main() {
	conf := ghrn.Config{}

	flag.StringVar(&conf.Org, "org", conf.Org, "Organization. (Required)")
	flag.StringVar(&conf.Repo, "repo", conf.Repo, "Repo. (Required)")
	flag.IntVar(&conf.StopAt, "stop-at", conf.StopAt, "PR number to stop at")
	flag.BoolVar(&conf.IncludeCommits, "include-commits", conf.IncludeCommits, "Include commit messages")
	flag.BoolVar(&conf.StopAtLatestRelease, "since-latest-release", conf.StopAtLatestRelease, "Stop at latest release's commit")
	flag.Parse()

	if conf.Org == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "\nError: -org is required.")
		os.Exit(1)
	}
	if conf.Repo == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "\nError: -repo is required.")
		os.Exit(1)
	}

	conf.GitHubToken = os.Getenv("GITHUB_TOKEN")

	ctx := context.Background()
	err := ghrn.BuildReleaseNotes(ctx, os.Stdout, conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		os.Exit(1)
	}
}
