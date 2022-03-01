// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package templates

import (
	"html/template"
	"io"
	"time"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_template_writer.go TemplateWriter

// TemplateWriter is an interface over html/template that's used to enable
// mocking.
type TemplateWriter interface {
	// Execute applies a parsed template to the specified data object,
	// writing the output to wr.
	Execute(wr io.Writer, data interface{}) error
}

// LockIndexData holds the fields needed to display the index view for locks.
type LockIndexData struct {
	LockPath      string
	RepoFullName  string
	PullNum       int
	Path          string
	Workspace     string
	Time          time.Time
	TimeFormatted string
}

// ApplyLockData holds the fields to display in the index view
type ApplyLockData struct {
	Locked        bool
	Time          time.Time
	TimeFormatted string
}

// IndexData holds the data for rendering the index page
type IndexData struct {
	Locks           []LockIndexData
	ApplyLock       ApplyLockData
	AtlantisVersion string
	// CleanedBasePath is the path Atlantis is accessible at externally. If
	// not using a path-based proxy, this will be an empty string. Never ends
	// in a '/' (hence "cleaned").
	CleanedBasePath string
}

var IndexTemplate = template.Must(template.New("index.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>atlantis</title>
  <meta name="description" content="">
  <meta name="author" content="">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <script src="{{ .CleanedBasePath }}/static/js/jquery-3.5.1.min.js"></script>
  <script>
    $(document).ready(function () {
      $("p.js-discard-success").toggle(document.URL.indexOf("discard=true") !== -1);
    });
    setTimeout(function() {
        $("p.js-discard-success").fadeOut('slow');
    }, 5000); // <-- time in milliseconds
  </script>
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/normalize.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/skeleton.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/custom.css">
  <link rel="icon" type="image/png" href="{{ .CleanedBasePath }}/static/images/atlantis-icon.png">
</head>
<body>
<div class="container">
  <section class="header">
    <a title="atlantis" href="{{ .CleanedBasePath }}/"><img class="hero" src="{{ .CleanedBasePath }}/static/images/atlantis-icon_512.png"/></a>
    <p class="title-heading">atlantis</p>
    <p class="js-discard-success"><strong>Plan discarded and unlocked!</strong></p>
  </section>
  <section>
    {{ if .ApplyLock.Locked }}
    <div class="twelve center columns">
      <h6><strong>Apply commands are disabled globally</strong></h6>
      <h6><code>Lock Status</code>: <strong>Active</strong></h6>
      <h6><code>Active Since</code>: <strong>{{ .ApplyLock.TimeFormatted }}</strong></h6>
      <a class="button button-primary" id="applyUnlockPrompt">Enable Apply Commands</a>
    </div>
    {{ else }}
    <div class="twelve columns">
      <h6><strong>Apply commands are enabled</strong></h6>
      <a class="button button-primary" id="applyLockPrompt">Disable Apply Commands</a>
    </div>
    {{ end }}
  </section>
  <br>
  <br>
  <br>
  <section>
    <p class="title-heading small"><strong>Locks</strong></p>
    {{ if .Locks }}
    {{ $basePath := .CleanedBasePath }}
    {{ range .Locks }}
      <a href="{{ $basePath }}{{.LockPath}}">
        <div class="twelve columns button content lock-row">
        <div class="list-title">{{.RepoFullName}} <span class="heading-font-size">#{{.PullNum}}</span> <code>{{.Path}}</code> <code>{{.Workspace}}</code></div>
        <div class="list-status"><code>Locked</code></div>
        <div class="list-timestamp"><span class="heading-font-size">{{.TimeFormatted}}</span></div>
        </div>
      </a>
    {{ end }}
    {{ else }}
    <p class="placeholder">No locks found.</p>
    {{ end }}
  </section>
  <div id="applyLockMessageModal" class="modal">
    <!-- Modal content -->
    <div class="modal-content">
      <div class="modal-header">
        <span class="close">&times;</span>
      </div>
      <div class="modal-body">
        <p><strong>Are you sure you want to create a global apply lock? It will disable applies globally</strong></p>
        <input class="button-primary" id="applyLockYes" type="submit" value="Yes">
        <input type="button" class="cancel" value="Cancel">
      </div>
    </div>
  </div>
  <div id="applyUnlockMessageModal" class="modal">
    <!-- Modal content -->
    <div class="modal-content">
      <div class="modal-header">
        <span class="close">&times;</span>
      </div>
      <div class="modal-body">
        <p><strong>Are you sure you want to release global apply lock?</strong></p>
        <input class="button-primary" id="applyUnlockYes" type="submit" value="Yes">
        <input type="button" class="cancel" value="Cancel">
      </div>
    </div>
  </div>
</div>
<footer>
v{{ .AtlantisVersion }}
</footer>
<script>

  function applyLockModalSetup(lockOrUnlock) {
      // Get the modal
      switch( lockOrUnlock ) {
      case "lock":
          var modal = $("#applyLockMessageModal");

          var btn = $("#applyLockPrompt");

          $("#applyLockYes").click(function() {
            $.ajax({
                url: '{{ .CleanedBasePath }}/apply/lock',
                type: 'POST',
                success: function(result) {
                  window.location.replace("{{ .CleanedBasePath }}/?discard=true");
                }
            });
          });

          break;
      case "unlock":
          var modal = $("#applyUnlockMessageModal");

          var btn = $("#applyUnlockPrompt");
          var btnApplyUnlock =

          $("#applyUnlockYes").click(function() {
            $.ajax({
                url: '{{ .CleanedBasePath }}/apply/unlock',
                type: 'DELETE',
                success: function(result) {
                  window.location.replace("{{ .CleanedBasePath }}/?discard=true");
                }
            });
          });

          break;
      default:
          throw("unsupported command " + lockOrUnlock)
      }

      return [modal, btn];
  }

  {{ if .ApplyLock.Locked }}
  var [modal, btn] = applyLockModalSetup("unlock");
  {{ else }}
  var [modal, btn] = applyLockModalSetup("lock");
  {{ end }}

  // Get the <span> element that closes the modal
  // using document.getElementsByClassName since jquery $("close") doesn't seem to work for btn click events
  var span = document.getElementsByClassName("close")[0];
  var cancelBtn = document.getElementsByClassName("cancel")[0];

  // When the user clicks the button, open the modal
  btn.click(function() {
    modal.css("display", "block");
  });

  // When the user clicks on <span> (x), close the modal
  span.onclick = function() {
    modal.css("display", "none");
  }
  cancelBtn.onclick = function() {
    modal.css("display", "none");
  }

  // When the user clicks anywhere outside of the modal, close it
  window.onclick = function(event) {
      if (event.target == modal) {
          modal.css("display", "none");
      }
  }
</script>
</body>
</html>
`))

// LockDetailData holds the fields needed to display the lock detail view.
type LockDetailData struct {
	LockKeyEncoded  string
	LockKey         string
	RepoOwner       string
	RepoName        string
	PullRequestLink string
	LockedBy        string
	Workspace       string
	Time            time.Time
	AtlantisVersion string
	// CleanedBasePath is the path Atlantis is accessible at externally. If
	// not using a path-based proxy, this will be an empty string. Never ends
	// in a '/' (hence "cleaned").
	CleanedBasePath string
}

var LockTemplate = template.Must(template.New("lock.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>atlantis</title>
  <meta name="description" content="">
  <meta name="author" content="">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/normalize.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/skeleton.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/custom.css">
  <link rel="icon" type="image/png" href="{{ .CleanedBasePath }}/static/images/atlantis-icon.png">
  <script src="{{ .CleanedBasePath }}/static/js/jquery-3.5.1.min.js"></script>
</head>
<body>
  <div class="container">
    <section class="header">
    <a title="atlantis" href="{{ .CleanedBasePath }}/"><img class="hero" src="{{ .CleanedBasePath }}/static/images/atlantis-icon_512.png"/></a>
    <p class="title-heading">atlantis</p>
    <p class="title-heading"><strong>{{.LockKey}}</strong> <code>Locked</code></p>
    </section>
    <div class="navbar-spacer"></div>
    <br>
    <section>
      <div class="eight columns">
        <h6><code>Repo Owner</code>: <strong>{{.RepoOwner}}</strong></h6>
        <h6><code>Repo Name</code>: <strong>{{.RepoName}}</strong></h6>
        <h6><code>Pull Request Link</code>: <a href="{{.PullRequestLink}}" target="_blank"><strong>{{.PullRequestLink}}</strong></a></h6>
        <h6><code>Locked By</code>: <strong>{{.LockedBy}}</strong></h6>
        <h6><code>Workspace</code>: <strong>{{.Workspace}}</strong></h6>
        <br>
      </div>
      <div class="four columns">
        <a class="button button-default" id="discardPlanUnlock">Discard Plan & Unlock</a>
      </div>
    </section>
  </div>
  <div id="discardMessageModal" class="modal">
    <!-- Modal content -->
    <div class="modal-content">
      <div class="modal-header">
        <span class="close">&times;</span>
      </div>
      <div class="modal-body">
        <p><strong>Are you sure you want to discard the plan and unlock?</strong></p>
        <input class="button-primary" id="discardYes" type="submit" value="Yes" data="{{.LockKeyEncoded}}">
        <input type="button" class="cancel" value="Cancel">
      </div>
    </div>
  </div>
<footer>
v{{ .AtlantisVersion }}
</footer>
<script>
  // Get the modal
  var modal = $("#discardMessageModal");

  // Get the button that opens the modal
  var btn = $("#discardPlanUnlock");
  var btnDiscard = $("#discardYes");
  var lockId = btnDiscard.attr('data');

  // Get the <span> element that closes the modal
  // using document.getElementsByClassName since jquery $("close") doesn't seem to work for btn click events
  var span = document.getElementsByClassName("close")[0];
  var cancelBtn = document.getElementsByClassName("cancel")[0];

  // When the user clicks the button, open the modal
  btn.click(function() {
    modal.css("display", "block");
  });

  // When the user clicks on <span> (x), close the modal
  span.onclick = function() {
    modal.css("display", "none");
  }
  cancelBtn.onclick = function() {
    modal.css("display", "none");
  }

  btnDiscard.click(function() {
    $.ajax({
        url: '{{ .CleanedBasePath }}/locks?id='+lockId,
        type: 'DELETE',
        success: function(result) {
          window.location.replace("{{ .CleanedBasePath }}/?discard=true");
        }
    });
  });

  // When the user clicks anywhere outside of the modal, close it
  window.onclick = function(event) {
      if (event.target == modal) {
          modal.css("display", "none");
      }
  }
</script>
</body>
</html>
`))

// ProjectJobData holds the data needed to stream the current PR information
type ProjectJobData struct {
	AtlantisVersion string
	ProjectPath     string
	CleanedBasePath string
}

var ProjectJobsTemplate = template.Must(template.New("blank.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>atlantis</title>
    <meta name="description" content>
    <meta name="author" content>
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/xterm.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/normalize.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/skeleton.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/custom.css">
    <link rel="icon" type="image/png" href="{{ .CleanedBasePath }}/static/images/atlantis-icon.png">
    <style>
      #terminal {
        position: fixed;
        top: 200px;
        left: 0px;
        bottom: 0px;
        right: 0px;
        overflow: auto;
        border: 5px solid white;
        }

      .terminal.xterm {
        padding: 10px;
      }
    </style>
  </head>

  <body>
    <section class="header">
    <a title="atlantis" href="{{ .CleanedBasePath }}"><img class="hero" src="{{ .CleanedBasePath }}/static/images/atlantis-icon_512.png"/></a>
    <p class="title-heading">atlantis</p>
    <p class="title-heading"><strong></strong></p>
    </section>
    <div class="spacer"></div>
    <br>
    <section>
      <div id="terminal"></div>
    </section>
  </div>
  <footer>
  </footer>

    <script src="{{ .CleanedBasePath }}/static/js/jquery-3.5.1.min.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-4.9.0.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-addon-attach-0.6.0.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-addon-fit-0.4.0.js"></script>

    <script>
      var term = new Terminal({scrollback: 15000});
      var socket = new WebSocket(
        (document.location.protocol === "http:" ? "ws://" : "wss://") +
        document.location.host +
        document.location.pathname +
        "/ws");
      window.addEventListener("unload", function(event) {
        websocket.close();
      })
      var attachAddon = new AttachAddon.AttachAddon(socket);
      var fitAddon = new FitAddon.FitAddon();
      term.loadAddon(attachAddon);
      term.loadAddon(fitAddon);
      term.open(document.getElementById("terminal"));
      fitAddon.fit();
      window.addEventListener("resize", () => fitAddon.fit());
    </script>
  </body>
</html>
`))

type ProjectJobsError struct {
	AtlantisVersion string
	ProjectPath     string
	CleanedBasePath string
}

var ProjectJobsErrorTemplate = template.Must(template.New("blank.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>atlantis</title>
    <meta name="description" content>
    <meta name="author" content>
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/xterm.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/normalize.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/skeleton.css">
    <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/custom.css">
    <link rel="icon" type="image/png" href="{{ .CleanedBasePath }}/static/images/atlantis-icon.png">
    <style>
      #terminal {
        width: 100%;
        height: 100%;
      }
    </style>
  </head>

  <body>
    <div class="container">
      <section class="header">
      <a title="atlantis" href="{{ .CleanedBasePath }}"><img class="hero" src="{{ .CleanedBasePath }}/static/images/atlantis-icon_512.png"/></a>
      <p class="title-heading">atlantis</p>
      <p class="title-heading"><strong></strong></p>
      </section>
      <div class="spacer"></div>
      <br>
      <section>
        <div id="terminal"></div>
      </section>
    </div>
    <footer>
    </footer>

    <script src="{{ .CleanedBasePath }}/static/js/jquery-3.5.1.min.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-4.9.0.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-addon-attach-0.6.0.js"></script>
    <script src="{{ .CleanedBasePath }}/static/js/xterm-addon-fit-0.4.0.js"></script>

    <script>
      var term = new Terminal();
      var socket = new WebSocket(
        (document.location.protocol === "http:" ? "ws://" : "wss://") + 
        document.location.host +
        document.location.pathname +
        "/ws");
      var attachAddon = new AttachAddon.AttachAddon(socket);
      var fitAddon = new FitAddon.FitAddon();
      term.loadAddon(attachAddon);
      term.loadAddon(fitAddon);
      term.open(document.getElementById("terminal"));
      term.write('Project Does Not Exist in PR')
      fitAddon.fit();
      window.addEventListener("resize", () => fitAddon.fit());
    </script>
  </body>
</html>
`))

// GithubSetupData holds the data for rendering the github app setup page
type GithubSetupData struct {
	Target        string
	Manifest      string
	ID            int64
	Key           string
	WebhookSecret string
	URL           string
}

var GithubAppSetupTemplate = template.Must(template.New("github-app.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>atlantis</title>
  <meta name="description" content="">
  <meta name="author" content="">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/normalize.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/skeleton.css">
  <link rel="stylesheet" href="{{ .CleanedBasePath }}/static/css/custom.css">
  <style>

    form {
      width: 100%;
    }

    form button {
      float: right;
    }

    textarea {
      width: 100%;
      height: 300px;
      font-family: monospace;
    }

    .config {
      display: flex;
      flex-direction: row;
      align-items: baseline;
      border-bottom: 1px solid #eee;
    }


    .config strong {
      width: 15%;
    }

    pre {
      background-color: #eee;
      padding: .5em;
      width: 80%;
    }
  </style>
  <link rel="icon" type="image/png" href="{{ .CleanedBasePath }}/static/images/atlantis-icon.png">
  <script src="{{ .CleanedBasePath }}/static/js/jquery-3.5.1.min.js"></script>
</head>
<body>
<div class="container">
  <section class="header">
    <a title="atlantis" href="{{ .CleanedBasePath }}"><img class="hero" src="{{ .CleanedBasePath }}/static/images/atlantis-icon_512.png"/></a>
    <p class="title-heading">atlantis</p>

    <p class="js-discard-success"><strong>
    {{ if .Target }}
      Create a github app
    {{ else }}
      Github app created successfully!
    {{ end }}
    </strong></p>
  </section>
  <section>
    {{ if .Target }}
    <form action="{{ .Target }}" method="POST">
      <textarea name="manifest">{{ .Manifest }}</textarea>
      <button type="submit">Setup</button>
    </form>
    {{ else }}
      <p>Visit <a href="{{ .URL }}/installations/new" target="_blank">{{ .URL }}/installations/new</a> to install the app for your user or organization, then <strong>update the following values</strong> in your config and <strong>restart Atlantis<strong>:</p>

      <ul>
        <li class="config"><strong>gh-app-id:</strong> <pre>{{ .ID }}</pre></li>
        <li class="config"><strong>gh-app-key-file:</strong> <pre>{{ .Key }}</pre></li>
        <li class="config"><strong>gh-webhook-secret:</strong> <pre>{{ .WebhookSecret }}</pre></li>
      </ul>
    {{ end }}
  </section>
</div>
</body>
</html>
`))
