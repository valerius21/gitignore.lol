package utils

type RepoInfo struct {
	URL       string
	LocalPath string
}

var DefaultRepoInfo RepoInfo

func init() {
	DefaultRepoInfo = RepoInfo{
		URL:       "https://github.com/github/gitignore.git",
		LocalPath: "/tmp/gitignore",
	}
}
