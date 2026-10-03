# Beta dashboard

Use **Try the beta dashboard** on the classic Atlantis dashboard, or visit
`<atlantis-url>/beta`. The classic dashboard remains available at `/` and the
beta includes a link back to it. No new server flag or frontend build is needed.

The beta groups tracked pull/merge requests by repository and workspace, with
project job output and active locks. It does not enumerate all open VCS requests.
Requests with missing workspace metadata, including workflow hooks, are kept
in an explicitly unavailable workspace group rather than assigned to `default`.

Search supports partial names and segment-anchored abbreviations across
repositories, workspaces and known lock owners. Space-separated terms and exact
dropdown filters combine with AND. User metadata may be unavailable for jobs
without locks. Request counts are unique by repository and request number,
even when a request appears in several workspaces.

Request headings use the stored PR/MR URL when available. Lock rows retain their
project-lock detail link and offer a separate PR/MR link. No provider URL guessing
or VCS API lookup is performed; missing request URLs remain plain text.

Global apply controls use the existing confirmation flow and controllers and
are hidden when global apply locking is disabled. Web authentication, terminal
streaming and existing lock/job routes retain their current behavior. Links and
assets honor the configured Atlantis base path.

## Integration boundary and follow-ups

The beta has a separate Go template, stylesheet and filtering script. The
classic and beta handlers share existing dashboard reads and HTTP 503 responses
for unavailable lock backends. Grouping is a pure view transformation; it does
not change persisted data, job identity, permissions or command execution.

Running/success/failure badges are not yet available. A follow-up needs an
authoritative latest-operation result and an active-job snapshot before
aggregating projects into request status. A closed output stream alone must not
imply success. Request titles and complete author metadata also need a safe data
source before being shown for every request.

The local fixture preview is maintained separately from the production feature.
Before upstream submission, agree on the rollout in an accepted issue and follow
the repository's AI usage policy and DCO requirements.
