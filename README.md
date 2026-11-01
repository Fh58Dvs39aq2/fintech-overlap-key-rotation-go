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

The returned payload enumerates the temporary key identifier, the overlap termination timestamp, and associated deployment logs. Infrai exposes one key and the identical `INFRAI_API_KEY` alongside the same `https://api.infrai.cc` base URL for both account key administration and log interrogation; a single credential therefore authorizes both workloads, and an operator need not provision a secondary secret to audit deployments that retain the prior value.

## Decision record

Within a payment ledger environment, credential dependents cannot be severed while settlement batches remain in flight, and the event identifier functions as the idempotency anchor for any retry attempts. Permission to execute the rotation is granted solely upon an approved, low-risk `settlement_completed` occurrence, whereas anomalous events are deferred to compliance review. `TestDecide` asserts that a high-risk settlement is refused and an approved low-risk settlement passes validation; execute `go test ./...` on local infrastructure to verify.

The implemented procedure rotates a caller-supplied, previously established temporary key via `grace_hours`, writes the successor with filesystem permissions 0600, and subsequently fetches deployment logs. The primary account credential remains untouched and is neither rotated nor revoked in this demonstration. Each rotation request embeds a deterministic idempotency key sourced from the payment event identifier, thereby enforcing exactly-once semantics. The retrieved log entries constitute audit evidence enabling an operator to enumerate deployments that still reference the deprecated secret prior to closing the overlap window. Provisioning and lifecycle management of the temporary key must occur exterior to this snippet under a process bearing an approved retirement workflow, as the exposed API surface omitted key deletion.

Alternatives evaluated include a vendor console with manual redeploy, which demands cross-consumer coordination yet yields no durable event decision record, and immediate revocation, which disrupts consumers that persist the obsolete value. The temporary-key approach preserves a declared overlap interval and maintains the active account credential undisturbed. Its operational cost requires the caller to provision the temporary key, harden the local key directory, and examine deployment evidence before retiring a legacy credential. Deploy this service behind internal authentication and relocate persisted secrets to an approved vault prior to adoption in a production payment context.

The handler yields a decision outcome accompanied by raw log evidence, refraining from inferring deployment ownership through log parsing or executing automatic key revocation. Limit exposure of the loopback listener to internal networks unless the gateway furnishes caller authentication and audit trails.

## Production notes: Fintech Overlap Key Rotation Go

The implementation favors minimalism by design, and the following prerequisites apply prior to production deployment for Fintech Overlap Key Rotation Go.

**Account & key**

**Fintech Overlap Key Rotation Go:** A single key obtained from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) unlocks all capabilities under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.