package gitrepository

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// enums: download options
const (
	DownloadNone        = -1
	DownloadFullPackage = iota
	DownloadPartialPackage
	DownloadSingleFile
	DownloadCustomPackage
)

const (
	DirectionNone = 0
	DirectionUp   = iota
	DirectionDown
)

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

func NewGitRepository(tempDir, ssid, rawUrl, branch string) *GitRepository {
	return &GitRepository{
		TempDir:      tempDir,
		SSID:         ssid,
		Url:          "",
		RawUrl:       rawUrl,
		CloneUrl:     "",
		RemoteUrl:    "",
		QueryUrl:     "",
		DirPath:      "",
		IsFile:       false,
		Protocol:     "",
		Scheme:       "",
		Hostname:     "",
		RawPath:      "",
		Path:         "",
		Owner:        "",
		Name:         "",
		DummyBranch:  "gitd-branch",
		Branch:       branch,
		ArchiveUrl:   "",
		FileUrl:      "",
		DownloadType: -1,
	}
}

// activate debug mode
func (r *GitRepository) ActivateDebugMode() {
	r.debugMode = true
}

// activate debug mode
func (r *GitRepository) isDebugModeActive() bool {
	return r.debugMode
}

// generate repository dir path
func (r *GitRepository) GetDirPath() string {
	branch := r.DummyBranch
	if r.Branch != "" {
		branch = r.Branch
	}

	return filepath.Join(r.TempDir, r.SSID, "repository", r.Owner, r.Name, branch)
}

// only https url accepted
/*
https://github.com/<owner>/<repo>
https://github.com/<owner>/<repo>.git -> .git remove
https://github.com/<owner>/<repo>/tree/<branch>/lib -> folder
https://github.com/<owner>/<repo>/blob/<branch>/lib/filesaver.min.js -> single file
https://github.com/<owner>/<repo>/blob/<branch>/internal/url/url.go#L20 -> #L20 removes
https://github.com/<owner>/<repo>/blob/<branch>/internal/url/url.go?deneme=12&obaraks=noway#L20 -> ?deneme=12&obaraks=noway#L20 remove

https://gitlab.com/<owner>/<repo>
https://gitlab.com/<owner>/<repo>.git
https://gitlab.com/<owner>/<repo>/-/tree/<branch>/config -> folder
https://gitlab.com/<owner>/<repo>/-/blob/<branch>/config/knative/api_groups.yml -> single file
https://gitlab.com/<owner>/<repo>/tree/<branch>/config -> folder
https://gitlab.com/<owner>/<repo>/blob/<branch>/config/knative/api_groups.yml -> single file

https://bitbucket.org/<owner>/<repo>.git
https://bitbucket.org/<owner>/<repo>/src/<branch>/ <- https://bitbucket.org/tiagoharris/url-shortener redirect
https://bitbucket.org/<owner>/<repo>/src/<branch>/cmd/ -> folder
https://bitbucket.org/<owner>/<repo>/src/<branch>/cmd/main.go -> single file


Supported: https://github.com/cli/cli/tree/marwan/localcs/api -> branch: marwan/localcs -> how to split this?

TODO: Url and RawUrl are the same? Why?
*/
func (r *GitRepository) Parse(sub string, direction int, filename string) error {

	// first
	re := regexp.MustCompile(`(?s)/-/`)
	r.RawUrl = re.ReplaceAllString(r.RawUrl, "/")
	//r.RawUrl = strings.Replace(r.RawUrl, "/-/", "/", 1) // fix: gitlab strange url
	if r.isDebugModeActive() {
		fmt.Println(r.RawUrl, "filename", filename)
	}

	// has to download single file
	// update raw url
	if filename != "" {
		re2 := regexp.MustCompile(`(?s)/tree/`)
		r.RawUrl = re2.ReplaceAllString(r.RawUrl, "/blob/")

		if r.isDebugModeActive() {
			fmt.Println(r.RawUrl, "filename", filename)
		}
	}

	// parse url
	u, err := url.Parse(r.RawUrl)
	if err != nil {
		return err
	}

	// find protocol
	r.Protocol = "https"

	// set scheme
	r.Scheme = u.Scheme

	// set hostname - not host
	r.Hostname = u.Hostname()

	// set path before clear unwanted querystring, fragments
	r.RawPath = filepath.Join(u.Path, filename)
	r.RawPath = strings.Replace(r.RawPath, u.RawFragment, "", 1)
	r.RawPath = strings.Replace(r.RawPath, u.RawQuery, "", 1)
	//r.RawPath = strings.Replace(r.RawPath, "/-/", "/", 1) // fix: gitlab strange url
	r.IsFile = !strings.HasSuffix(r.RawUrl, "/")   // only useful for bitbucket.org url
	r.RawPath = strings.TrimSuffix(r.RawPath, "/") // remove last slashes

	if r.isDebugModeActive() {
		fmt.Println(r.RawPath)
	}

	// repeater counter
	repeater := strings.Count(r.RawPath, "/")
	if r.isDebugModeActive() {
		fmt.Println(repeater, "repeater")
	}
	if repeater < 2 {
		return errors.New("not valid git url")
	}

	// multi slashes branch name
	branchNameRepeater := 0
	if r.Branch != "" {
		branchNameRepeater = strings.Count(r.Branch, "/")
	}

	// n[1] = owner, n[2] = repo, n[3] = tree|blob, n[4] = branch, n[5] = ../../../...
	n := strings.SplitN(r.RawPath, "/", 6+branchNameRepeater) // fixed n times all urls
	r.Owner = n[1]
	r.Name = n[2]

	if r.isDebugModeActive() {
		fmt.Println("split n: ", n, "branchNameRepeater", branchNameRepeater, "sub", sub)
	}

	if strings.HasSuffix(r.Name, ".git") {
		r.Name = strings.Replace(r.Name, ".git", "", 1)
		r.RawPath = strings.Replace(r.RawPath, ".git", "", 1)
	}

	if repeater >= 3 {
		if n[3] == "blob" || n[3] == "tree" || n[3] == "src" {
			if branchNameRepeater > 0 {
				// branch name contains slash
				if len(n) > (4 + branchNameRepeater + 1) {
					r.Path = n[4+branchNameRepeater+1]
				}
			} else {
				r.Branch = n[4]
				if len(n) > 5 {
					r.Path = n[5]
				}
			}

			// Bug and TODO
			// Bitbucket.org url has src not tree or blob.
			// if url not slashes, after download system failed because IsFile value not correct
			// r.IsFile = !strings.HasSuffix(r.Path, "/")
			if n[3] == "tree" {
				r.IsFile = false
			} else if n[3] == "blob" {
				r.IsFile = true
			}
		} else {
			return errors.New("not valid git branch")
		}
	} else {
		r.IsFile = false
	}

	// sub folder calculation for jump between folders
	if sub == "root" {
		// clone url must be return: jump to root folder
		if r.Path != "" {
			r.RawPath = strings.Replace(r.RawPath, r.Path, "", 1)
			r.Path = ""
		}
	} else if sub != "" {
		if r.Path != "" {
			index := -1
			if direction == DirectionUp {
				if strings.Count(r.Path, sub) == 1 {
					index = strings.LastIndex(r.Path, sub)
				} else {
					index = strings.Index(r.Path, sub)
				}
			}

			if index == -1 {
				r.Path = filepath.Join(r.Path, sub)
				r.RawPath = filepath.Join(r.RawPath, sub)
			} else {
				r.Path = r.Path[0 : index+len(sub)]

				rawIndex := strings.Index(r.RawPath, sub)
				r.RawPath = r.RawPath[0 : rawIndex+len(sub)]
			}
		} else {
			r.Path = sub
			r.RawPath = filepath.Join(r.RawPath, r.Path)
		}
	}

	// generate real url
	r.CloneUrl = r.Scheme + "://" + r.Hostname + "/" + r.Owner + "/" + r.Name + ".git"
	r.RemoteUrl = "git@" + r.Hostname + ":" + r.Owner + "/" + r.Name + ".git"
	r.Url = r.Scheme + "://" + r.Hostname + r.RawPath

	// generate pathDir
	r.DirPath = r.GetDirPath()

	// Generate Remote Url Addresses
	r.ArchiveUrl = r.getArchiveUrl()
	r.FileUrl = r.getFileUrl("[PATH]")
	r.QueryUrl = r.GetQueryUrl(r.Path)

	// Download Type
	if r.CloneUrl == r.Url+".git" {
		// full package
		r.DownloadType = DownloadFullPackage
	} else if r.Path == "" {
		// full package
		r.DownloadType = DownloadFullPackage
	} else if r.IsFile {
		// single file
		r.DownloadType = DownloadSingleFile
	} else {
		// partial package
		r.DownloadType = DownloadPartialPackage
	}

	if r.isDebugModeActive() {
		fmt.Printf("%#v\n", r)
	}
	return nil
}

func (r *GitRepository) WithoutCloneUrl() string {
	return strings.Replace(r.CloneUrl, ".git", "", 1)
}

func (r *GitRepository) UpdateBranch(branch string) {
	r.Branch = branch

	// Generate Remote Url Addresses
	r.ArchiveUrl = r.getArchiveUrl()
	r.FileUrl = r.getFileUrl("[PATH]")
}

// generate archive url
// Add: is multiple slash branch name, slashes removes
func (r *GitRepository) getArchiveUrl() string {
	switch r.Hostname {
	case "gitlab.com":
		// https://[HOSTNAME]/[OWNER]/[NAME]/-/archive/[BRANCH]/gitlab-[BRANCH].[EXT]
		return fmt.Sprintf("https://%s/%s/%s/-/archive/%s/gitlab-%s.%s", r.Hostname, r.Owner, r.Name, r.Branch, strings.ReplaceAll(r.Branch, "/", "-"), "zip")
	case "github.com":
		// https://[HOSTNAME]/[OWNER]/[NAME]/archive/refs/heads/[BRANCH].[EXT]
		// github archive url redirect always
		// TODO: Redirect to https://codeload.github.com/[OWNER]/[NAME]/zip/refs/heads/[BRANCH]
		return fmt.Sprintf("https://%s/%s/%s/archive/refs/heads/%s.%s", r.Hostname, r.Owner, r.Name, r.Branch, "zip")
	case "bitbucket.org":
		// https://[HOSTNAME]/[OWNER]/[NAME]/get/[BRANCH].[EXT]
		return fmt.Sprintf("https://%s/%s/%s/get/%s.%s", r.Hostname, r.Owner, r.Name, r.Branch, "zip")
	}

	return ""
}

// generate file url
func (r *GitRepository) getFileUrl(path string) string {
	switch r.Hostname {
	case "gitlab.com":
		// https://[HOSTNAME]/[OWNER]/[NAME]/-/blob/[BRANCH]/[PATH]
		// https://gitlab.com/gitlab-org/gitlab/-/raw/dc-move-assignees-widget/.git-blame-ignore-revs
		return fmt.Sprintf("https://%s/%s/%s/-/raw/%s/%s", r.Hostname, r.Owner, r.Name, r.Branch, path)
	case "github.com":
		// https://[HOSTNAME]/[OWNER]/[NAME]/blob/[BRANCH]/[PATH]
		// https://raw.githubusercontent.com/101arrowz/fflate/master/.npmignore
		return fmt.Sprintf("https://%s/%s/%s/%s/%s", "raw.githubusercontent.com", r.Owner, r.Name, r.Branch, path)
	case "bitbucket.org":
		// https://[HOSTNAME]/[OWNER]/[NAME]/raw/[BRANCH]/[PATH]
		// https://bitbucket.org/micovery/sock-rpc/raw/v1.0.0/package.json
		return fmt.Sprintf("https://%s/%s/%s/raw/%s/%s", r.Hostname, r.Owner, r.Name, r.Branch, path)
	}

	return ""
}

// generate folder url
func (r *GitRepository) GetQueryUrl(path string) string {
	baseUrl := fmt.Sprintf("%s://%s/%s/%s", r.Scheme, r.Hostname, r.Owner, r.Name)

	if r.Branch != "" {
		if path != "" && r.IsFile {
			index := strings.LastIndex(r.Path, "/")
			if index != -1 {
				path = r.Path[0:index]
			} else {
				path = ""
			}
		}

		switch r.Hostname {
		case "gitlab.com":
			// https://[HOSTNAME]/[OWNER]/[NAME]/-/blob/[BRANCH]/[PATH]
			return fmt.Sprintf("%s/tree/%s/", baseUrl, filepath.Join(r.Branch, path))
		case "github.com":
			// https://[HOSTNAME]/[OWNER]/[NAME]/blob/[BRANCH]/[PATH]
			return fmt.Sprintf("%s/tree/%s/", baseUrl, filepath.Join(r.Branch, path))
		case "bitbucket.org":
			// https://[HOSTNAME]/[OWNER]/[NAME]/src/[BRANCH]/[PATH]
			return fmt.Sprintf("%s/src/%s/", baseUrl, filepath.Join(r.Branch, path))
		}
	}

	return baseUrl
}

// find real folder path
func (r *GitRepository) FindRealFolderPath(path string) string {
	return r._findRealFolderPath(path)
}
func (r *GitRepository) _findRealFolderPath(path string) string {
	if path != "" && r.IsFile {
		index := strings.LastIndex(r.Path, "/")
		if index != -1 {
			path = r.Path[0:index]
		} else {
			path = ""
		}
	}

	return path
}
