# GoGitViz
GoGitViz is a simple terminal tool that visualizes your local Git commit history in a GitHub-style contributions graph.

Unlike GitHub or GitLab, which only show commits pushed to their servers, GoGitViz scans your local repositories and builds a contributions heatmap based on your commits -- across all projects and remotes combined.

This means you get a complete picture of your coding activity, even if you:

- Work on both GitHub and GitLab
- Keep private or offline repositories
- Contribute across multiple accounts

With GoGitViz, your contribution graph reflects your true local commit history.

## Features
- Scan your machine for local Git repositories.
- Track commits by email.
- Generate a contributions graph for the past 6 months (configurable).
- Colored terminal heatmap like GitHub.

## Installation
```bash
git clone https://github.com/yourname/gogitlocalstats
cd gogitlocalstats
go mod tidy
go build -o gogitstats
```
## Usage

**Add a folder to scan**
```bash 
./gogitstats --add /path/to/code
```

**Generate stats**
```bash 
./gogitstats --email "you@example.com"
```

**Change time range (default 6 months)**
```bash 
./gogitstats --email "you@example.com" --months 12
```
