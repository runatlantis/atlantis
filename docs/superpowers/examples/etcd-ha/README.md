# Embedded etcd HA — reference Kubernetes manifests

Reference manifests for the dual-mode embedded etcd HA topology described in
[`../../specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md`](../../specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md).
These are examples for the `runatlantis/helm-charts` chart to adopt, not a
packaged chart. TLS secrets, RBAC identities, and image references are
placeholders.

Two logical replica types are built from the **same** Atlantis image:

| File | Role |
| --- | --- |
| `member-statefulset.yaml` | Fixed 3-member StatefulSet; each pod runs Atlantis **and** embeds one etcd voter. Stable identity + per-member PVC. |
| `peer-service.yaml` | Headless service publishing not-ready peer addresses so quorum can form. |
| `client-service.yaml` | Client-facing service selecting **both** member and client replicas. |
| `client-deployment.yaml` | Zero-or-more independently autoscaled client replicas in external etcd mode. |
| `pdb.yaml` | Pod disruption budget preserving quorum (`maxUnavailable: 1`). |

Key invariants (design §324):

- `podManagementPolicy: Parallel` — ordered readiness would deadlock quorum
  formation, because the first member cannot become ready until its peers start.
- `updateStrategy: OnDelete` — a voter-count or template change must never
  auto-restart voters; membership transitions are an explicit operator runbook.
- Anti-affinity + topology spread — voters land on distinct nodes/zones.
- `maxUnavailable: 1` for every supported voter count (3, 5, 7); the runbook also
  permits only one manual member deletion at a time.
- Startup probes allow the full `--etcd-startup-timeout` for quorum formation.
- Only the **client** Deployment is eligible for ordinary autoscaling; the member
  StatefulSet replica count is a quorum setting, never an HPA target.

Initial bring-up uses `--etcd-embedded-lifecycle=bootstrap
--etcd-embedded-startup-purpose=maintenance` (non-serving) until the namespace is
initialized, then members restart in `restart`/`serve`. See the design's
"Voter-count lifecycle" and "Migration and rollback" sections.
