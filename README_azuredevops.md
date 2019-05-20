# Atlantis

<p align="center">
  <img src="./runatlantis.io/.vuepress/public/hero.png" alt="Atlantis Logo"/><br><br>
  <b>Terraform Pull Request Automation</b>
</p>

## Azure Devops Support (alpha)

As of tag `v0.7.1` of [go-azuredevops](https://github.com/mcdafydd/go-azuredevops), basic suport for comments, statuses, running `atlantis help|plan|apply` appears to be functional.  There is still much to do in terms of writing tests and documentation and operational testing.

Known issues:

* Merging pull requests doesn't seem to be removing the locks

