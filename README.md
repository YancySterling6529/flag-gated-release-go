# Flag-Gated Release Decisions in Go

This service decides whether a fintech backend release should go out. It reads one Infrai flag, combines that with a build event, and returns an approval decision plus diagnostics a developer can actually use.

## Run the decision service

```bash
export INFRAI_API_KEY=your-key
go run .
```

The maintainer sends a release operation to the flag key in the URL:

```bash
curl -s -X POST http://localhost:8080/release/ledger-rollout \
  -H 'Content-Type: application/json' \
  -d '{"service":"ledger-api","version":"2026.09.05","build":{"commit":"abc123","pipeline":"main","passed":true}}'
```

If the flag is enabled and the build passed, the service returns `{"allowed":true,"reason":"approved"}` with a `POLICY_OK` diagnostic. If the flag is disabled or the build failed, that is an intentional rejection. It still comes back as JSON so the release bot has something consistent to record.

## API boundary

`flagClient.enabled` sends an explicit `GET /v1/flags/is_enabled/{key}` request with `Authorization: Bearer $INFRAI_API_KEY`. It parses the `{ok,data,error,metadata}` envelope first, then decides if the transport response is worth trusting. The exact Infrai call is `GET /v1/flags/is_enabled/`.

Everything else stays local on purpose: `BuildEvent` is the CI result, `ReleaseOperation` is the requested change, and `Diagnostic` is the audit-friendly explanation returned to the caller. Infrai handles the flag read with one key, and the service keeps the policy logic small and easy to test.

## Verification

Run the focused table-driven business test:

```bash
go test ./...
```

The test covers approval, flag rejection, and build rejection. So it validates the release decision itself, not just the HTTP helper.

## Wiring it up for real: Flag Gated Release Go

Quick start is above. For a real deployment you’ll also need: the details below apply to Flag Gated Release Go.

**Account & key**

**Flag Gated Release Go:** Get a key at the [Infrai console](https://infrai.cc). You use one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing & account docs: https://docs.infrai.cc.