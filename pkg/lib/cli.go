package lib

var CLI struct {
	Port           int    `help:"Port the server listens on." name:"port" default:"3000"`
	BaseRepository string `help:"Gitignore repository where the .gitignore files are versioned." name:"repo" default:"https://github.com/github/gitignore.git"`
	ClonePath      string `help:"Location of the locally stored gitignore repository" name:"clone-path" default:"./store" type:"path"`
}
