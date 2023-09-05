# Gitd - Git Parse Url

Parse git url simple way. (SCP-Style url not supported yet)

## Feature

- Use the same code of [Gitdownloadmanager Api Service](https://gitdownloadmanager.com)
- Generate Github, Bitbucket, Gitlab repository download full package url address
- Supports all git url address without scp-styles

## Git Repository

```go
// git repository
type GitRepository struct {
 TempDir string
 SSID    string

 debugMode bool

 Url       string // clean url after parse
 RawUrl    string // user set this dirty url
 CloneUrl  string
 RemoteUrl string // remote url for git git@github.com:username/repo.git
 QueryUrl  string // for search bar
 DirPath   string

 IsFile bool

 Protocol    string // https|ssh (scp-style - not supported yet)
 Scheme      string
 Hostname    string
 RawPath     string
 Path        string // file or folder path in this repository for download
 Owner       string
 Name        string // repository name - repo
 DummyBranch string // if branch name is empty, use this name
 Branch      string

 ArchiveUrl   string // download branch package
 FileUrl      string // download from single file url
 DownloadType int
}
```

## Example Use

simple parse action

```go
// args
tempDir := "" // only use gitdownloadmanager service
ssid := "" // only use gitdownloadmanager service for session
rawUrl := "https://github.com/cli/cli"
branch := "" // if has a multi slash branch name, you can set it

// create new repository obj
gitRepository := NewGitRepository(tempDir, ssid, rawUrl, branch)

// parse if you can
sub := "" // use only breadcrumb query for find root folder
direction := DirectionNone // use only breadcrumb query for which direction to go
filename := "" // use file user download for only one file
if err := gitRepository.Parse(sub, direction, filename); err != nil {
    fmt.Errorf("GitRepository.Parse() error = %#v", err)
}

fmt.Prinf("GitRepository.Parse() = %#v", gitRepository)
```

response

```go
GitRepository{
    TempDir:      "",
    SSID:         "",
    Url:          "https://github.com/cli/cli",
    RawUrl:       "https://github.com/cli/cli",
    CloneUrl:     "https://github.com/cli/cli.git",
    RemoteUrl:    "git@github.com:cli/cli.git",
    QueryUrl:     "https://github.com/cli/cli",
    DirPath:      "repository/cli/cli/gitd-branch",
    IsFile:       false,
    Protocol:     "https",
    Scheme:       "https",
    Hostname:     "github.com",
    RawPath:      "/cli/cli",
    Path:         "",
    Owner:        "cli",
    Name:         "cli",
    DummyBranch:  "gitd-branch",
    Branch:       "",
    ArchiveUrl:   "https://github.com/cli/cli/archive/refs/heads/.zip",
    FileUrl:      "https://raw.githubusercontent.com/cli/cli//[PATH]",
    DownloadType: 1,
}
```
