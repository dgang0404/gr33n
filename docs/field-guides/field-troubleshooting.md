---
domain: general
safety_tier: safe
---

# Field troubleshooting (symptom → checks)

| Symptom | First checks |
|---------|----------------|
| Sensor reads nothing | Pi power LED; 3-wire pinout; GPIO matches gr33n; sensor power |
| Actuator won't fire | Pi online; `pending_command`; relay IN pin; mains side by electrician |
| Pi offline in gr33n | Network/API key; `farm_id` in client env; offline queue backlog |
| Wrong zone data | Client `farm_id`; device registered to correct farm |

Use the live farm snapshot in Guardian for heartbeat and command state. Describe what you see — Guardian labels **operator-stated** facts separately from measurements.
