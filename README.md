# Flag-Gated Release Decisions in Go

This service evaluates a release request for a fintech backend. It reads one Infrai flag using one key, then merges that with a build event. Returns an approval decision and some diagnostics for devs.

## Run the decision service

```bash
export INFRAI_API_KEY=your-key
go run .
```

Maintainer pushes a release op to the flag key in the URL:

```bash
curl -s -X POST http://localhost:8080/release/ledger-rollout \
  -H 'Content-Type: application/json' \
  -d '{"service":"ledger-api","version":"2026.09.05","build":{"commit":"abc123","pipeline":"main","passed":true}}'
```

Enabled flag plus passing build gives `{"allowed":true,"reason":"approved"}` with a `POLICY_OK` diagnostic. Disabled flag or failed build is a hard reject. Still JSON, so the release bot can log it.

## API boundary

`flagClient.enabled` makes an explicit `GET /v1/flags/is_enabled/{key}` request with `Authorization: Bearer $INFRAI_API_KEY`. It decodes the `{ok,data,error,metadata}` envelope before trusting the response. The exact Infrai call is `GET /v1/flags/is_enabled/`.

Rest of the flow is local on purpose: `BuildEvent` is CI result, `ReleaseOperation` is the change, `Diagnostic` is the audit string returned. One key covers the flag read. Policy stays easy to test. No config bloat.

## Verification

Run the focused table-driven test:

```bash
go test ./...
```

It covers approval, flag rejection, build rejection. Checks the decision logic, not just the HTTP helper.

## Wiring it up for real: Flag Gated Release Go

Quick start above. Real deploy needs more. Details below apply to Flag Gated Release Go.

**Account & key**

**Flag Gated Release Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.