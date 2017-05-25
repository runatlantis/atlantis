package main

import (
	"html/template"
	"net/http"
	"strings"
	"github.com/elazarl/go-bindata-assetfs"
)

var indexTemplate = template.Must(template.New("index.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>atlantis</title>
  <meta name="description" content="">
  <meta name="author" content="">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/static/css/normalize.css">
  <link rel="stylesheet" href="/static/css/skeleton.css">
  <link rel="stylesheet" href="/static/css/custom.css">
  <link rel="icon" type="image/png" href="/static/atlantis-icon.png">
</head>
<body>
  <div class="container">
    <section class="header">
    <a title="atlantis" href="/"><img src="/static/atlantis-icon.png" /></a>
    <p style="font-family: monospace, monospace; font-size: 1.1em; text-align: center;">atlantis</p>
    </section>
    <nav class="navbar">
      <div class="container">
      </div>
    </nav>
    <div class="navbar-spacer"></div>
    <br>
    <section>
    <p style="font-family: monospace, monospace; font-size: 1.0em; text-align: center;"><strong>Environments</strong></p>
      {{ if . }}
        {{ range . }}
          <a href="/lock/view/{{.RepoOwner}}/{{.RepoName}}/{{.PullRequestId}}"><div class="twelve columns button content"><div class="list-title">{{.RepoOwner}}/{{.RepoName}} - <span class="heading-font-size">#{{.PullRequestId}}</span></div> <div class="list-status"><code>Locked</code></div> <div class="list-timestamp"><span class="heading-font-size">{{.Timestamp}}</span></div></div></a>
        {{ end }}
      {{ else }}
        <p class="placeholder">No environments found.</p>
      {{ end }}
    </section>
  </div>
</body>
</html>
`))

type binaryFileSystem struct {
	fs http.FileSystem
}

func NewBinaryFileSystem(root string) *binaryFileSystem {
	return &binaryFileSystem{
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: root},
	}
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.fs.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}
