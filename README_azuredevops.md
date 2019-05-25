# Atlantis

<p align="center">
  <img src="./runatlantis.io/.vuepress/public/hero.png" alt="Atlantis Logo"/><br><br>
  <b>Terraform Pull Request Automation</b>
</p>

## Azure Devops Support (alpha)

As of tag `v0.7.1` of [go-azuredevops](https://github.com/mcdafydd/go-azuredevops), basic suport for comments, statuses, running `atlantis help|plan|apply` appears to be functional.  There is still much to do in terms of writing tests and documentation and operational testing.

Known issues:

* Merging pull requests doesn't seem to be removing the locks

## Problems building in a fork

Since the parent project is using dep and the vendor folder, I'm currently developing this inside the following folder on my workstation:

`$GOPATH/src/github.com/runatlantis/atlantis/`

This deals with the problem of breaking local package imports by trying to build inside `$GOPATH/src/github.com/mcdafydd/atlantis`.  With Go modules, apparently `go mod replace` can help to address this issue more cleanly, but for now this works.  I don't want to do a global search replace and then change it back before submitting any pull requests. 

I'm happy to hear any suggestions about better ways to handle this!

