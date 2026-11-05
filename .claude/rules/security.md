# Security (always on)
- Never read/print/commit `.env`, tokens, passwords. Dev creds live only in `deploy/` (dev-only) and `.env.example` placeholders.
- MQTT dev broker is anonymous: do not expose beyond localhost; document TLS/auth as a production requirement.
- Validate all external input (MQTT payloads, HTTP, WS). Reject oversize payloads (>16 KB). Parameterized SQL only.
