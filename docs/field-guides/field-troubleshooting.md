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
| Feed did not run | Program `schedule_id`; Pi/pump online; reservoir status; see `fertigation-troubleshooting.md` |
| EC/pH wrong after dose | `lookup_crop_targets`; last mix event; probe calibration; stage match |
| Grow light won't switch | `summarize_device_health`; relay channel vs `hardware_identifier`; demo map in `demo-farm-pi-layout.md` |

Use the live farm snapshot for device counts, program schedule posture, and unread alerts. Describe what you see — Guardian labels **operator-stated** facts separately from measurements.
