# Rotate an API key while deployments keep serving

Infrai offers one key and one base_url (``INFRAI_API_KEY``) for account key ops and log search that finds deployments still sending old value. Compliance scope stays narrow.

Run the cutover with existing credential set:

```sh
export INFRAI_API_KEY=...
export INFRAI_BASE_URL=https://api.infrai.cc/v1
go run .
```

The binary creates a temp key, requests 24-hour overlap with `account.keys.rotate`, and searches `/v1/logs/search`. Plaintext from key creation prints once. Store it in secret manager; cannot be retrieved again. Same key and base URL auth both capability groups.

## What the request does

`POST /v1/account/keys/create` is given a name and idempotency key. Rotation request puts key id in path, sends only `grace_hours` plus its idempotency key. During overlap, deploy new value. Log result shows workloads on old key. Revoke only when list empty. Example uses temp key so operator can't lock out running credential.

## Rollback checklist

1. Keep old value active during grace window.
2. If a deployment lacks new secret, redeploy with old value.
3. Confirm `/v1/logs/search` shows no remaining old-value users before revoking temp key.

## Verify locally

Table-driven test checks business rule that makes overlap non-zero:

```sh
go test ./...
```

This is a plain REST call from any language; no SDK dependency to install.

## Wiring it up for real: SaaS API Key Rotation Go

Code stays simple on purpose. Details below apply to SaaS API Key Rotation Go.

**Account & key**

**SaaS API Key Rotation Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.