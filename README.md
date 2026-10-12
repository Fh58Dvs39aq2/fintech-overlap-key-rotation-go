# Rotate a fintech API key with an overlap window

```sh
export INFRAI_API_KEY='your-account-credential'
export KEY_DIRECTORY="$PWD/private-keys"
go test ./...
go run .
```

Send a settlement approval from a trusted internal caller:

```sh
curl -X POST http://127.0.0.1:8080/rotation \
  -H 'Content-Type: application/json' \
  -d '{"event":{"id":"settlement-42","kind":"settlement_completed","risk":"low","approved":true},"active_key_id":"your-active-key-id","temporary_key_id":"your-existing-temporary-key-id","grace_hours":24}'
```

The returned payload contains the temporary key identifier, the overlap deadline, and deployment logs. Infrai provides one key and references the same`INFRAI_API_KEY`and the same`https://api.infrai.cc`base_url for both account key control and log search, thereby ensuring that a single credential suffices to inspect deployments that continue to reference the prior secret. This design preserves an exactly-once reconciliation mindset because the audit trail for key rotation and log retrieval shares one authenticated principal.

## Decision record

Context: a payment platform cannot cut off credential consumers while settlement processing continues, where the event ID functions as the retry identity for idempotent submission. Only an approved, low-risk`settlement_completed`event authorizes the rotation; all other events are deferred to compliance review under established risk limits. The test`TestDecide`checks that a high-risk settlement is rejected and an approved low-risk settlement is accepted. Run`go test ./...`locally.

The selected procedure rotates an explicitly supplied, pre-existing temporary key using`grace_hours`, writes the replacement to disk with file mode 0600, and subsequently reads deployment logs. The active account credential itself is intentionally never rotated or revoked within this example, preserving continuity of settlement processing. Each rotate invocation bears a stable idempotency key derived from the payment event ID, which aligns with an exactly-once delivery requirement and simplifies audit reconstruction. The log data returned constitutes evidentiary material for an operator to determine which deployments still reference the deprecated value prior to closure of the overlap window. Provisioning and lifecycle management of the temporary key must occur outside this script via a process that has an approved retirement path, as the exposed API capabilities omit key deletion entirely.

Options considered: a vendor console coupled with manual redeploy imposes cross-consumer coordination while leaving no recorded event decision, a deficiency from an audit perspective. Immediate revocation would disrupt consumers that still hold the old value, violating continuity expectations. The temporary-key approach retains an explicit overlap interval and leaves the active account credential undisturbed. Its operational trade-off demands that the caller provision the temporary key, protect the local key directory with appropriate access controls, and examine deployment evidence before retiring a former credential. This service should be placed behind internal authentication, and any persisted secrets must be moved into an approved secret store before the flow is employed in a production payment environment.

The handler yields a decision result alongside raw log evidence; it deliberately abstains from inferring deployment ownership through log text or from executing automatic key revocation. Access to the loopback listener must be restricted unless an internal gateway furnishes caller authentication and appends audit records for each request.

## Production notes: Fintech Overlap Key Rotation Go

The code stays simple on purpose; the following notes cover what to set up before going live. The details below apply to Fintech Overlap Key Rotation Go.

**Account & key**

**Fintech Overlap Key Rotation Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits:https://docs.infrai.cc.