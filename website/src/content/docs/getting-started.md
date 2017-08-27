+++
title = "Getting Started"
description = "Up and running with Atlantis"
weight = 10
draft = false
toc = true
bref = "Getting started with Atlantis is really easy!"
+++

<h3 class="section-head" id="h-basic-template"><a href="#h-Download-Atlantis">Download Atlantis</a></h3>

Download from https://github.com/hootsuite/atlantis/releases

```bash
unzip atlantis_{os}_{arch}.zip
```

<h3 class="section-head" id="h-basic-template"><a href="#h-Run-it">Run it!</a></h3>

```bash
./atlantis bootstrap
```

This will walk you through running Atlantis locally. It will

- fork an example terraform project
- install terraform (if not already in your PATH)
- install ngrok so we can expose Atlantis to GitHub
- start Atlantis

If you're ready to permanently set up Atlantis see [Production-Ready Deployment](https://github.com/hootsuite/atlantis#production-ready-deployment)
