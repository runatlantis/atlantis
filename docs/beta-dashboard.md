# Beta dashboard

Use **Try the beta dashboard** on the classic Atlantis dashboard, or visit
`<atlantis-url>/beta`. The classic dashboard remains available at `/` and the
beta includes a link back to it. No new server flag or frontend build is needed.

The Jobs section groups tracked output by repository, workspace, and PR/MR.
Lock-only projects appear exclusively in Locks, with their owner and management
links. Locks may supply known owner and PR/MR metadata for existing Jobs entries,
but never create them. Empty job mappings are omitted. The beta does not enumerate
all open VCS requests.
Requests with missing workspace metadata, including workflow hooks, are kept
in an explicitly unavailable workspace group rather than assigned to `default`.

Search supports partial names and segment-anchored abbreviations across
repositories, workspaces and known lock owners. Space-separated terms and exact
dropdown filters combine with AND. User metadata may be unavailable for jobs
without locks. The Jobs badge, filter summary, and workspace counts reflect
individual tracked jobs, including workflow hooks, rather than request counts.
Job output is held in memory; after a restart, Jobs may be empty while persisted
Locks remain visible. This does not discard locks or plans.

Navigation and search/filter controls stay within the viewport while the results
scroll independently. The results region is keyboard-focusable. On short screens,
the filter panel can scroll separately so controls remain reachable without
covering the results.
Job and lock rows use horizontal space for their details, wrapping on
narrow screens. Repeated operation labels are omitted; each output link and its
timestamp remain available. Section navigation scrolls
the results panel; Apply controls brings its enable/disable button into view,
including when the panel is too short to show the entire section.
On desktop, the Atlantis version and build information stays in a separate
sidebar footer while navigation scrolls. The Classic dashboard link sits above
the build information, rather than in a page header. It remains available on
mobile, where only the build information is hidden.

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
