package server

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
	LockURL      string
	RepoFullName string
	PullNum      int
	Time         time.Time
}

// IndexData holds the data for rendering the index page
type IndexData struct {
	Locks           []LockIndexData
	AtlantisVersion string
}

var indexTemplate = template.Must(template.New("index.html.tmpl").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>atlantis</title>
  <meta name="description" content="">
  <meta name="author" content="">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <script src="/static/js/jquery-3.2.1.min.js"></script>
  <script>
    $(document).ready(function () {
      $("p.js-discard-success").toggle(document.URL.indexOf("discard=true") !== -1);
    });
    setTimeout(function() {
        $("p.js-discard-success").fadeOut('slow');
    }, 5000); // <-- time in milliseconds
  </script>
  <link rel="stylesheet" href="/static/css/normalize.css">
  <link rel="stylesheet" href="/static/css/skeleton.css">
  <link rel="stylesheet" href="/static/css/custom.css">
  <link rel="icon" type="image/png" href="/static/images/atlantis-icon.png">
</head>
<body>
<div class="container">
  <section class="header">
    <a title="atlantis" href="/"><img src="/static/images/atlantis-icon.png"/></a>
    <p class="title-heading">atlantis</p>
    <p class="js-discard-success"><strong>Plan discarded and unlocked!</strong></p>
  </section>
  <nav class="navbar">
    <div class="container">
    </div>
  </nav>
  <div class="navbar-spacer"></div>
  <br>
  <section>
    <p class="title-heading small"><strong>Locks</strong></p>
    {{ if .Locks }}
    {{ range .Locks }}
      <a href="{{.LockURL}}">
        <div class="twelve columns button content lock-row">
        <div class="list-title">{{.RepoFullName}} - <span class="heading-font-size">#{{.PullNum}}</span></div>
        <div class="list-status"><code>Locked</code></div>
        <div class="list-timestamp"><span class="heading-font-size">{{.Time}}</span></div>
        </div>
      </a>
    {{ end }}
    {{ else }}
    <p class="placeholder">No locks found.</p>
    {{ end }}
  </section>
</div>
</body>
</html>
`))

// LockDetailData holds the fields needed to display the lock detail view.
type LockDetailData struct {
	UnlockURL       string
	LockKeyEncoded  string
	LockKey         string
	RepoOwner       string
	RepoName        string
	PullRequestLink string
	LockedBy        string
	Workspace       string
	Time            time.Time
}

var lockTemplate = template.Must(template.New("lock.html.tmpl").Parse(`
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
  <link rel="icon" type="image/png" href="/static/images/atlantis-icon.png">
  <script src="/static/js/jquery-3.2.1.min.js"></script>
</head>
<body>
  <div class="container">
    <section class="header">
    <a title="atlantis" href="/"><img src="/static/images/atlantis-icon.png"/></a>
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
        url: '/locks?id='+lockId,
        type: 'DELETE',
        success: function(result) {
          window.location.replace("/?discard=true");
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
