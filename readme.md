`github-release-notes` is a basic utility for generating release notes content from GitHub Pull Request history.

## Usage

Download a binary from the [releases](https://github.com/buchanae/github-release-notes/releases) page, or run `go get github.com/buchanae/github-release-notes`.

Run:
```
github-release-notes -org ohsu-comp-bio -repo funnel
- PR #519 webdash: fixed elapsedTime calculation
- PR #516 storage/swift: wrap errors with useful context
- PR #515 Moving code
- PR #514 build: fix release notes command
- PR #513 Webdash upgrades
- PR #512 worker/docker: log container metadata
- PR #511 Unexport
- PR #510 build: goreleaser, 0.6.0, github release notes gen
...
```

You can stop generating notes at a specific PR:
```
github-release-notes -org ohsu-comp-bio -repo funnel -stop-at 513
- PR #519 webdash: fixed elapsedTime calculation
- PR #516 storage/swift: wrap errors with useful context
- PR #515 Moving code
- PR #514 build: fix release notes command
```

You can generating notes with PR that not included latest release:
```
github-release-notes -org ohsu-comp-bio -repo funnel -since-latest-release 
- PR #594 cmd/worker: run task from file
- PR #593 storage/ftp: add FTP support
```

You can include the git commit messages for each PR:
```
github-release-notes -org ohsu-comp-bio -repo funnel -include-commits
- PR #519 webdash: fixed elapsedTime calculation
    - 7675a5a5d577340b47e4dbdc5b83338c35a26392 webdash: fixed elapsedTime calculation

- PR #516 storage/swift: wrap errors with useful context
    - 53b583c71da5e06c7dddd26e480f9099d6e8e60d storage/swift: wrap errors with useful context
```

You can use an [API access token][tok] by setting the `GITHUB_TOKEN` environment variable:
```
export GITHUB_TOKEN=1234...
github-release-notes -org ohsu-comp-bio -repo funnel
```

[tok]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
