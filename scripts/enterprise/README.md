# Enterprise deployment helpers (community extension)

**Status:** Placeholder — no required scripts ship with the core platform yet.

Large multi-site operators (see [`docs/hypothetical-enterprise-topology.md`](../../docs/hypothetical-enterprise-topology.md)) often need **repeatable bring-up**:

- Bulk farm / zone creation from YAML  
- Pi `config.yaml` generation from a device manifest  
- Commons catalog pack import across many `farm_id`s  
- Post-deploy smoke (health, one reading, one actuator round-trip)

## Contributing

If you build deployment pipeline tooling against the **public HTTP API**:

1. Prefer **config + scripts that call the API** over forking `cmd/api` unless you must patch core behavior.  
2. Open a **pull request** to this directory with a short README per tool (inputs, outputs, idempotency story).  
3. Do not commit secrets, `.env` files, or customer-specific hostnames.

## License note

gr33n platform code is **[AGPL v3](../../LICENSE)**. Ops scripts in this folder are intended as **operator tooling**; if your organization modifies the gr33n **application** itself and exposes it to users over a network, AGPL obligations apply to that software — consult counsel. Upstreaming deployment helpers here benefits everyone and avoids fork drift.

## Related

- [`docs/hypothetical-enterprise-topology.md`](../../docs/hypothetical-enterprise-topology.md)  
- [`docs/plans/phase_30_guardian_change_requests.plan.md`](../../docs/plans/phase_30_guardian_change_requests.plan.md)  
- [`docs/plans/phase_31_field_validation_and_edge.plan.md`](../../docs/plans/phase_31_field_validation_and_edge.plan.md)
