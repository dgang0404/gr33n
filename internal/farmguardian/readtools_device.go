// Phase 65 WS1 — Pi & hardware diagnostics read tool (summarize_device_health).

package farmguardian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/hardware"
)

const (
	deviceHealthOfflineAfter        = 5 * time.Minute
	deviceHealthConfigStaleAfter    = 30 * time.Minute
	deviceHealthSensorStaleFallback = 15 * time.Minute
	deviceHealthMaxDevices          = 4
)

// DeviceHealthGroundingRule reminds Guardian to use platform wiring records.
const DeviceHealthGroundingRule = `Device wiring (Phase 65): Use summarize_device_health for Pi heartbeat, config sync age, sensor GPIO/source, actuator relay channel, reading freshness, and GPIO conflicts. Platform wiring is operator-entered — caveat that physical wires may differ. Never walk through mains AC troubleshooting.`

var (
	summarizeDeviceHealthIntent = regexp.MustCompile(`(?i)\b(offline|not connecting|heartbeat|wiring|gpio|channel|relay|stuck|not updating|not responding|wrong reading|impossible value|device health|config sync|stale reading|pi is|why is .+ showing)\b|\b(sensor|temp|humidity|co2|ec)\b.{0,40}\b(stuck|wrong|not updating|frozen)\b|\b(fan|pump|light|actuator)\b.{0,40}\b(not responding|won't|wont|not working)\b`)
	sensorRoutePattern          = regexp.MustCompile(`^/sensors/(\d+)`)
)

func shouldRunSummarizeDeviceHealthReadIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if deviceHealthRouteContext(ref) {
		if q == "" || summarizeDeviceHealthIntent.MatchString(q) {
			return true
		}
	}
	if ref != nil && strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
		if summarizeDeviceHealthIntent.MatchString(q) {
			return true
		}
	}
	return summarizeDeviceHealthIntent.MatchString(q)
}

func deviceHealthRouteContext(ref *ContextRef) bool {
	if ref == nil {
		return false
	}
	if !strings.EqualFold(ref.Type, "route") {
		return false
	}
	path := strings.TrimSpace(ref.Path)
	switch {
	case path == "/pi-setup", path == "/sensors", path == "/actuators":
		return true
	case strings.HasPrefix(path, "/sensors/"):
		return true
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/devices/new"):
		return true
	default:
		return false
	}
}

func renderSummarizeDeviceHealth(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	devices, err := resolveDevicesForHealth(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "summarize_device_health: no edge devices found for this farm — register a Pi in the device wizard first.", nil
	}

	var blocks []string
	for i, dev := range devices {
		if i >= deviceHealthMaxDevices {
			blocks = append(blocks, fmt.Sprintf("(+%d more devices not listed)", len(devices)-deviceHealthMaxDevices))
			break
		}
		block, berr := renderSingleDeviceHealth(ctx, q, farmID, dev)
		if berr != nil {
			return "", berr
		}
		if block != "" {
			blocks = append(blocks, block)
		}
	}
	if len(blocks) == 0 {
		return "summarize_device_health: no device health data available.", nil
	}
	return strings.Join(blocks, "\n\n"), nil
}

func resolveDevicesForHealth(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) ([]db.Gr33ncoreDevice, error) {
	all, err := q.ListDevicesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	active := make([]db.Gr33ncoreDevice, 0, len(all))
	for _, d := range all {
		if d.DeletedAt.Valid {
			continue
		}
		active = append(active, d)
	}
	if len(active) == 0 {
		return nil, nil
	}

	lower := strings.ToLower(question)
	for _, d := range active {
		if d.DeviceUid != nil && strings.Contains(lower, strings.ToLower(*d.DeviceUid)) {
			return []db.Gr33ncoreDevice{d}, nil
		}
		if strings.Contains(lower, strings.ToLower(d.Name)) {
			return []db.Gr33ncoreDevice{d}, nil
		}
	}

	if ref != nil && strings.EqualFold(ref.Type, "route") {
		if m := sensorRoutePattern.FindStringSubmatch(strings.TrimSpace(ref.Path)); len(m) == 2 {
			if sid, perr := strconv.ParseInt(m[1], 10, 64); perr == nil {
				if s, serr := q.GetSensorByID(ctx, sid); serr == nil && s.FarmID == farmID && s.DeviceID != nil {
					if dev, derr := q.GetDeviceByID(ctx, *s.DeviceID); derr == nil && dev.FarmID == farmID {
						return []db.Gr33ncoreDevice{dev}, nil
					}
				}
			}
		}
	}

	if sensors, serr := q.ListSensorsByFarm(ctx, farmID); serr == nil {
		for _, s := range sensors {
			if s.Name != "" && strings.Contains(lower, strings.ToLower(s.Name)) && s.DeviceID != nil {
				if dev, derr := q.GetDeviceByID(ctx, *s.DeviceID); derr == nil && dev.FarmID == farmID {
					return []db.Gr33ncoreDevice{dev}, nil
				}
			}
		}
	}

	if ref != nil && strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
		zid := ref.ID
		if zoneDevs, zerr := q.ListDevicesByZone(ctx, &zid); zerr == nil && len(zoneDevs) > 0 {
			out := make([]db.Gr33ncoreDevice, 0, len(zoneDevs))
			for _, d := range zoneDevs {
				if !d.DeletedAt.Valid {
					out = append(out, d)
				}
			}
			if len(out) > 0 {
				return out, nil
			}
		}
	}

	sort.Slice(active, func(i, j int) bool { return active[i].Name < active[j].Name })
	return active, nil
}

func renderSingleDeviceHealth(ctx context.Context, q db.Querier, farmID int64, dev db.Gr33ncoreDevice) (string, error) {
	uid := ""
	if dev.DeviceUid != nil {
		uid = *dev.DeviceUid
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_device_health — %s", strings.TrimSpace(dev.Name)))
	if uid != "" {
		b.WriteString(fmt.Sprintf(" (%s)", uid))
	}
	b.WriteByte('\n')

	status := string(dev.Status)
	hbLine := "last heartbeat: never"
	if dev.LastHeartbeat.Valid {
		age := timeSince(dev.LastHeartbeat.Time)
		hbLine = fmt.Sprintf("last heartbeat %s ago", humanizeAge(age))
		if age > deviceHealthOfflineAfter {
			status += " ⚠ OFFLINE"
		}
	} else {
		status += " ⚠ OFFLINE"
	}

	cfgAge, cfgOK := deviceConfigSyncAge(dev.Config)
	cfgLine := "config synced: never fetched"
	if cfgOK {
		cfgLine = fmt.Sprintf("config synced %s ago", humanizeAge(cfgAge))
		if cfgAge > deviceHealthConfigStaleAfter {
			cfgLine += " ⚠ STALE"
		}
	}
	b.WriteString(fmt.Sprintf("Status: %s · %s · %s (v%d)", status, hbLine, cfgLine, dev.ConfigVersion))

	deviceID := dev.ID
	sensors, err := q.ListSensorsByDevice(ctx, &deviceID)
	if err != nil {
		return "", err
	}
	actuators, err := q.ListActuatorsByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}
	var devActuators []db.Gr33ncoreActuator
	for _, a := range actuators {
		if a.DeletedAt.Valid {
			continue
		}
		if a.DeviceID != nil && *a.DeviceID == dev.ID {
			devActuators = append(devActuators, a)
		}
	}

	sort.Slice(sensors, func(i, j int) bool { return sensors[i].Name < sensors[j].Name })
	sort.Slice(devActuators, func(i, j int) bool { return devActuators[i].Name < devActuators[j].Name })

	b.WriteString(fmt.Sprintf("\nSensors (%d):", len(sensors)))
	if len(sensors) == 0 {
		b.WriteString("\n- (none assigned to this device)")
	} else {
		gpioGroups := groupSensorsByGPIO(sensors)
		for _, s := range sensors {
			b.WriteByte('\n')
			b.WriteString(renderSensorHealthLine(ctx, q, s, gpioGroups))
		}
	}

	b.WriteString(fmt.Sprintf("\nActuators (%d):", len(devActuators)))
	if len(devActuators) == 0 {
		b.WriteString("\n- (none assigned to this device)")
	} else {
		for _, a := range devActuators {
			b.WriteByte('\n')
			b.WriteString(renderActuatorHealthLine(ctx, q, a))
		}
	}

	return b.String(), nil
}

func renderSensorHealthLine(ctx context.Context, q db.Querier, s db.Gr33ncoreSensor, gpioGroups map[int][]string) string {
	w, _ := hardware.ExtractWiring(s.Config)
	wiringLabel := hardware.FormatLabel(w)
	if wiringLabel == "" {
		wiringLabel = "no wiring on file"
	}

	readingAge := "no readings yet"
	stale := false
	reading, rerr := q.GetLatestReadingBySensor(ctx, s.ID)
	if rerr == nil {
		age := timeSince(reading.ReadingTime)
		readingAge = humanizeAge(age) + " ago"
		stale = sensorReadingStale(age, s.ReadingIntervalSeconds)
	} else if !errors.Is(rerr, pgx.ErrNoRows) {
		readingAge = "lookup error"
	}

	line := fmt.Sprintf("- %s: %s · last reading %s", strings.TrimSpace(s.Name), wiringLabel, readingAge)
	if stale {
		line += " ⚠ STALE"
	}
	if w != nil && w.GPIOPin != nil {
		if peers := gpioGroups[*w.GPIOPin]; len(peers) > 1 {
			line += fmt.Sprintf(" (shares GPIO %d with %s)", *w.GPIOPin, strings.Join(peers, ", "))
		}
	}
	return line
}

func renderActuatorHealthLine(ctx context.Context, q db.Querier, a db.Gr33ncoreActuator) string {
	channel := "no channel on file"
	if a.HardwareIdentifier != nil && strings.TrimSpace(*a.HardwareIdentifier) != "" {
		channel = formatRelayChannelLabel(*a.HardwareIdentifier)
	} else if w, _ := hardware.ExtractWiring(a.Config); w != nil {
		channel = hardware.FormatLabel(w)
	}

	zoneLabel := ""
	if a.ZoneID != nil {
		if z, err := q.GetZoneByID(ctx, *a.ZoneID); err == nil {
			zoneLabel = fmt.Sprintf(" · zone: %s", z.Name)
		}
	}
	state := ""
	if a.CurrentStateText != nil && strings.TrimSpace(*a.CurrentStateText) != "" {
		state = fmt.Sprintf(" · state: %s", strings.TrimSpace(*a.CurrentStateText))
	}
	return fmt.Sprintf("- %s: %s%s%s", strings.TrimSpace(a.Name), channel, zoneLabel, state)
}

func groupSensorsByGPIO(sensors []db.Gr33ncoreSensor) map[int][]string {
	out := make(map[int][]string)
	for _, s := range sensors {
		w, _ := hardware.ExtractWiring(s.Config)
		if w == nil || w.GPIOPin == nil {
			continue
		}
		pin := *w.GPIOPin
		out[pin] = append(out[pin], strings.TrimSpace(s.Name))
	}
	return out
}

func sensorReadingStale(age time.Duration, intervalSec *int32) bool {
	if intervalSec != nil && *intervalSec > 0 {
		return age > time.Duration(*intervalSec)*3*time.Second
	}
	return age > deviceHealthSensorStaleFallback
}

func deviceConfigSyncAge(config json.RawMessage) (time.Duration, bool) {
	if len(config) == 0 {
		return 0, false
	}
	var root struct {
		LastConfigFetchAt string `json:"last_config_fetch_at"`
	}
	if err := json.Unmarshal(config, &root); err != nil {
		return 0, false
	}
	ts := strings.TrimSpace(root.LastConfigFetchAt)
	if ts == "" {
		return 0, false
	}
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, ts)
	}
	if err != nil {
		return 0, false
	}
	return timeSince(parsed), true
}

func formatRelayChannelLabel(hi string) string {
	ch, err := strconv.Atoi(strings.TrimSpace(hi))
	if err != nil {
		return "relay HAT ch " + strings.TrimSpace(hi)
	}
	return fmt.Sprintf("relay HAT ch %d (stack %d, relay %d)", ch, ch/8, (ch%8)+1)
}
