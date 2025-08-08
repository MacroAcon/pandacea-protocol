# Pandacea Protocol: 4 Sprint Production Readiness Plan (v1.0)

## Purpose

This plan turns the current end to end demo into a production ready protocol. It breaks the work into four sprints with clear goals, scope, deliverables, acceptance criteria, risks, and artifacts to commit.

**Guardrails for all sprints**

- No breaking changes without a version bump and migration notes.
- All new code ships with unit tests, integration tests, and docs updates.
- Logs follow the structured logging spec and never contain PII.
- Every PR must pass `make verify` locally and in CI.

---

## Sprint 1: Replace mocks and freeze interfaces

**Goal**: Remove training mocks, wire real PySyft, and freeze schemas and events so downstream code can rely on them.

### Scope

1. **PySyft worker and privacy accountant**

- Implement a Python worker that performs local DP training using PySyft (DP SGD or equivalent) on a small reference dataset.
- Add a simple privacy accountant that tracks per user epsilon consumption. Respect a per user budget. Fail closed if budget is exceeded.
- Persist DP parameters per job: epsilon, clip, noise multiplier, seed, accountant state pointer.
- Support `MOCK_DP=1` as a development fallback that keeps the same artifact schema.

2. **Artifact and consent manifest schema freeze**

- Define and lock JSON schemas for the model artifact and consent manifest.
- Hash the consent manifest and reference the hash on chain in the lease flow.
- Record artifact integrity using SHA 256 and store the hash in the manifest.

3. **Event and API freeze**

- Freeze the event set for contracts used by indexers. Write an event catalog with field names and types.
- Freeze the agent REST API surface used by the SDK. Version it as v1.

4. **Contracts testing uplift**

- Add Foundry invariant tests for dispute staking, reputation decay, and settlement flows.
- Add fuzz tests for parameter edges and unexpected state transitions.

### Deliverables

- `agent-backend/worker/train_worker.py` with real PySyft path and mock fallback.
- `docs/PRIVACY_BOUNDARY.md` with data flow diagrams and DP parameter notes.
- `docs/schemas/artifact.schema.json` and `docs/schemas/consent.schema.json`.
- `docs/events/catalog.md` that lists all contract events with topics and meanings.
- Invariant and fuzz test suites with a short README that explains what they cover.

### Acceptance criteria

- `make demo-real` runs end to end with real PySyft locally and produces an artifact that matches the schema and includes epsilon, clip, noise.
- Privacy accountant blocks any job that would exceed the configured budget and returns a clear error.
- Contract invariant tests pass locally and in CI. Minimum 95 percent line coverage for affected contracts.
- API v1 and event catalog are published and tagged in the repo.

### Risks and mitigations

- **DP instability on Windows**: pin versions and add a compatibility note in README. Provide a Docker path as a fallback.
- **Schema churn**: use JSON schema versioning. Mark v1 as frozen at sprint close.

### Artifacts to commit

- Worker code, schemas, accountant module, tests, docs, and updated Make targets.

---

## Sprint 2: Security and economics

**Goal**: Harden the agent against abuse and prove the economic design holds up under adversarial behavior.

### Scope

1. **Networking and DoS controls**

- Bind agent identities to libp2p keys that map to on chain addresses.
- Require authenticated peers for sensitive routes. Add token bucket rate limiting per identity and per IP. Add global backpressure when CPU or memory is high.
- Add concurrency caps per identity. Add greylisting and temporary bans on repeated abuse.
- Validate message sizes and schema before processing. Reject oversize payloads early.

2. **Adversarial simulation pack**

- Build a simulation suite that includes honest earners, colluding cohorts, griefers, and hoarders.
- Model the latest economic rules. Run parameter sweeps for dispute stake levels and decay rates.
- Produce a report with plots. Show bounded griefing cost, Sybil resistance assumptions, and sensitivity analysis.

3. **External audit for contracts**

- Prepare a narrow scope for the first audit: PGT, LeaseAgreement, reputation and dispute logic.
- Run the audit, triage findings, patch, and re test. Document the before and after.

### Deliverables

- `docs/security/agent_abuse_controls.md` with the exact limits, thresholds, and error codes.
- `sims/` folder with code, seeds, and notebooks that reproduce figures.
- `docs/economics/simulation_report.md` with a narrative and key plots.
- Audit report and remediation notes checked into `docs/audits/`.

### Acceptance criteria

- Load test shows the agent stays responsive under configured limits and sheds load safely.
- Simulation report demonstrates that honest strategies dominate over time within the chosen parameter ranges.
- All audit findings resolved or accepted with a documented rationale.

### Risks and mitigations

- **False positives from rate limits**: add allow lists for staging and CI. Tune with load tests.
- **Simulation mismatch**: align on chain constants and simulation parameters through a single config source of truth.

### Artifacts to commit

- Abuse controls code, tests, simulation code, report, and audit materials.

---

## Sprint 3: Ops and SDK stability

**Goal**: Make the system observable, reproducible, and stable. Freeze API v1 and ship SDKs with a compatibility matrix.

### Scope

1. **Observability**

- Add metrics and traces using OpenTelemetry. Expose Prometheus endpoints.
- Adopt a structured logging spec with field names, levels, and correlation IDs.
- Create Grafana dashboards for latency, error rates, DP job outcomes, and chain calls.

2. **Reproducible builds and supply chain**

- Produce SBOMs for agent and SDK builds. Sign container images. Publish provenance.
- Add deterministic build flags and pinned versions. Document the steps.

3. **API v1 freeze and SDK semver**

- Freeze the REST API and Python and Go SDK interfaces. Start semantic versioning.
- Add a compatibility test matrix that runs SDK tests across API versions in CI.

4. **Resilience and recovery drills**

- Write a disaster recovery plan with RPO and RTO targets.
- Run a backup restore drill and a dependency outage drill.

### Deliverables

- `docs/observability/runbook.md` and dashboards in `ops/grafana/`.
- SBOMs in `dist/sbom/` and signed images published to the registry.
- SDK version 1.0.0 for Python and Go with change logs and migration notes.
- `docs/ops/dr_plan.md` with the drill results and gaps.

### Acceptance criteria

- A fresh operator can deploy staging using the runbook and get green dashboards in under one hour.
- CI publishes SBOMs and signed images for every tag.
- Compatibility tests pass for the current and previous minor API versions.
- DR drill restores staging within the target RTO and no data loss beyond the target RPO.

### Risks and mitigations

- **Metrics cost or noise**: set clear retention and sampling policies. Keep only useful signals.
- **SDK fragmentation**: enforce CI gates for backward compatibility and document deprecations early.

### Artifacts to commit

- Observability code, dashboards, SBOMs, signing config, SDK releases, and docs.

---

## Sprint 4: Compliance and launch

**Goal**: Ship policies, runbooks, and staging soak. Lock upgrade safety. Complete the Go or No Go review and then launch.

### Scope

1. **Threat model and incident response**

- Document threats using a simple STRIDE style and privacy risks using a LINDDUN style checklist.
- Write an incident response runbook with roles, escalation paths, and timelines.
- Add an abuse handling workflow for spam, fraud, and DMCA style complaints.

2. **Policies and user processes**

- Publish Terms of Service and Privacy Policy that match how the system works.
- Implement data access and deletion request handling. Test it on staging.

3. **Staging soak on Polygon Amoy**

- Deploy to Amoy with monitoring and alerting. Use synthetic traffic and limited partner traffic.
- Run a two week soak with rate limits, quotas, and backpressure enabled. Track stability and costs.

4. **Upgrade safety and launch hygiene**

- Put contracts behind a timelock and a 3 of 5 multisig with clear roles.
- Run the final audit sign off. Freeze the event catalog and API v1.
- Prepare the Launch Runbook and the Go or No Go checklist.

### Deliverables

- `docs/security/threat_model.md` and `docs/security/incident_response.md`.
- Published ToS and Privacy Policy in `docs/policy/`.
- Amoy staging deployment notes and dashboards.
- `docs/launch/runbook.md` and `docs/launch/go_no_go_checklist.md`.

### Acceptance criteria

- A privacy and security review signs off on the threat model and runbooks.
- Data access and deletion requests are processed within the stated timelines during a staging drill.
- Staging soak completes with error rates and resource usage inside targets.
- Contracts are controlled by the timelock and multisig. Final audit report is published with all findings addressed.

### Risks and mitigations

- **Policy drift from implementation**: run a review with engineers and counsel. Update both sides together.
- **Soak fails**: capture issues, fix, and extend soak until stability targets are met.

### Artifacts to commit

- Threat model, policies, runbooks, staging results, and launch docs.

---

## Program timeline and tracking

- Suggested cadence: 2 week sprints with a hard review at the end of each sprint.
- Use milestone boards that mirror this document. Each card links to code, tests, and docs.
- Maintain a risk log with owner, impact, likelihood, and mitigation.

## Definition of done for the program

All items below must be true to call the protocol production ready.

- External contract audit complete and remediated. Timelock and multisig live.
- API v1 and event catalog frozen and published. Compatibility tests green.
- Privacy accountant enforced. Consent manifests hashed on chain. Deletion flow exercised.
- Agent abuse controls active and tested. Load tests show stable operation.
- Observability, SBOMs, signed images, and DR plan in place and proven.
- Staging soak complete with success metrics inside targets.
- Go or No Go checklist fully green.

## Owners and next steps

- Assign a DRI for each sprint and a backup.
- Open issues for each deliverable in this plan. Tag them with the sprint label.
- Start Sprint 1 by opening a tracking issue that links to all child tasks. Create a checklist and close the sprint only when the acceptance criteria are met.

## Appendix A: Example JSON schemas (shortened)

**Artifact schema**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Artifact",
  "type": "object",
  "required": ["job_id", "model", "epsilon", "dp", "hash", "created_at"],
  "properties": {
    "job_id": {"type": "string"},
    "model": {"type": "string"},
    "epsilon": {"type": "number"},
    "accuracy": {"type": "number"},
    "n": {"type": "integer"},
    "dp": {
      "type": "object",
      "required": ["enabled", "epsilon", "clip", "noise_multiplier"],
      "properties": {
        "enabled": {"type": "boolean"},
        "epsilon": {"type": "number"},
        "clip": {"type": "number"},
        "noise_multiplier": {"type": "number"}
      }
    },
    "hash": {"type": "string"},
    "seed": {"type": "integer"},
    "created_at": {"type": "string", "format": "date-time"}
  }
}
```

**Consent manifest schema**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ConsentManifest",
  "type": "object",
  "required": ["subject_id", "data_product_id", "terms", "hash"],
  "properties": {
    "subject_id": {"type": "string"},
    "data_product_id": {"type": "string"},
    "terms": {
      "type": "object",
      "required": ["purpose", "retention", "revocation"],
      "properties": {
        "purpose": {"type": "string"},
        "retention": {"type": "string"},
        "revocation": {"type": "string"}
      }
    },
    "hash": {"type": "string"}
  }
}
```

## Appendix B: Standard checklists

**Per sprint exit checklist**

-

**Go or No Go checklist items**

-

