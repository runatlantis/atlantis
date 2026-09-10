# External etcd HA — reference Kubernetes manifests (phase 1)

Reference manifests for running Atlantis active-active against an **external**
etcd cluster, per
[`../../specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md`](../../specs/2026-09-07-dual-mode-embedded-etcd-ha-design.md).
These are examples for the `runatlantis/helm-charts` chart to adopt, not a
packaged chart. TLS secrets, RBAC identities, and image references are
placeholders.

**Prerequisite:** a production external etcd cluster (bring-your-own, or a
separate etcd StatefulSet) reachable over mTLS. The embedded in-process etcd
voter topology — the member StatefulSet, headless peer service, and quorum PDB —
is **deferred to phase 2** and is not included here.

| File | Role |
| --- | --- |
| `client-deployment.yaml` | Atlantis in external etcd mode. Autoscalable (0..N via HPA); every replica accepts ingress and owner-routes internally. |
| `client-service.yaml` | Client-facing service selecting the Atlantis replicas. |
| `pdb.yaml` | Pod disruption budget keeping at least one Atlantis replica available. |

Key points:

- Selecting `--locking-db-type=etcd --etcd-mode=external` always activates
  active-active PR ownership and owner-routing; there is no etcd-backed
  multi-replica mode without fencing (design "Ownership and routing").
- Atlantis replicas do **not** participate in Raft. Scaling them up or down never
  changes etcd membership; membership is a property of the external cluster.
- Point `--etcd-endpoints` at your external etcd client endpoints (all `https`,
  hostname-verified). Production requires a complete client TLS identity
  (`--etcd-ca-file`, `--etcd-cert-file`, `--etcd-key-file`).
- The internal command transport requires a shared token
  (`--internal-command-token-file`) and CA (`--internal-command-ca-file`); the
  advertise allowlist (`--replica-advertise-allowlist`) is an SSRF guard on
  owner forwarding.
- `/readyz` reports backend authority plus a live ownership session, so a replica
  that loses quorum or its lease is removed from the Service.

Losing etcd quorum makes every replica unready and fails new commands closed by
design. See the design's "Failure semantics" and "Migration and rollback"
sections.
