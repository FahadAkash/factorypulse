# MQTT topics & delivery (contract v1)
| Topic | QoS | Retain | Producer → Consumer |
|---|---|---|---|
| `nova/{factory}/{line}/{machine}/state` | 1 | **yes** (last state) | Unity → Go |
| `nova/{factory}/{line}/{machine}/production` | 1 | no | Unity → Go |
| `nova/{factory}/{line}/{machine}/heartbeat` | 0 | no | Unity → Go |
| `nova/{factory}/{line}/{machine}/sensor` | 0 | no | Unity → Go |
| `nova/control/sim` | 1 | no | Go/API → Unity (`start`,`stop`,`set_rate`,`fail_machine`,`replay`) |
| `nova/control/ack` | 1 | no | Unity → Go |

- Payload = envelope from `event-schema.json`, UTF-8 JSON. Go consumes with shared subscription `$share/fp/nova/+/+/+/#`.
- Delivery is **at-least-once** → duplicates are expected; Go dedupes on `event_id`.
- Breaking change ⇒ bump `schema_version`, support N and N-1 in Go for one release.
