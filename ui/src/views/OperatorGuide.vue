<template>
  <div class="p-4 sm:p-6 max-w-3xl mx-auto space-y-10 pb-24 md:pb-10">
    <header class="space-y-2">
      <h1 class="text-2xl font-bold text-green-400 flex items-center gap-2">
        Operator guide
        <HelpTip position="bottom">
          Same mental model as <strong>docs/operator-tour.md</strong> in the repo. Use this page when you want terms and a click path without leaving the app.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500 leading-relaxed">
        Recommended order for new farms — each step opens in the app.
      </p>
    </header>

    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-3">
      <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Suggested walk</h2>
      <ol class="list-decimal list-inside space-y-2 text-sm text-zinc-300">
        <li><router-link class="text-gr33n-400 hover:underline" to="/">Farm home</router-link> — context, counts, quick links</li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/zones">Zones</router-link> — grow areas before attaching gear</li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/sensors">Sensors</router-link> · <router-link class="text-gr33n-400 hover:underline" to="/actuators">Controls</router-link> · <router-link class="text-gr33n-400 hover:underline" to="/setpoints">Setpoints</router-link></li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/schedules">Schedules</router-link> · <router-link class="text-gr33n-400 hover:underline" to="/automation">Rules</router-link></li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/tasks">Tasks</router-link></li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/fertigation">Fertigation</router-link></li>
        <li><router-link class="text-gr33n-400 hover:underline" to="/chat">Farm Guardian</router-link> — optional AI; change requests need Confirm (see glossary)</li>
      </ol>
      <p class="text-xs text-zinc-600 pt-2">
        Also: <router-link class="text-gr33n-500 hover:underline" to="/alerts">Alerts</router-link>,
        <router-link class="text-gr33n-500 hover:underline" to="/guardian/requests">Guardian requests</router-link>,
        <router-link class="text-gr33n-500 hover:underline" to="/costs">Costs</router-link>,
        <router-link class="text-gr33n-500 hover:underline" to="/farm-knowledge">Knowledge</router-link> (RAG).
      </p>
    </section>

    <section class="space-y-4">
      <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Glossary</h2>
      <p class="text-xs text-zinc-600">
        Stable wording for UI copy and farm-scoped help (aligns with <code class="text-zinc-500">workflow-guide.md</code> §11).
      </p>
      <dl class="space-y-4">
        <div v-for="entry in glossary" :key="entry.term" class="bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3">
          <dt class="text-gr33n-400 font-semibold text-sm">{{ entry.term }}</dt>
          <dd class="text-sm text-zinc-400 mt-1 leading-relaxed">{{ entry.body }}</dd>
        </div>
      </dl>
    </section>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950/50 px-4 py-3 text-xs text-zinc-500 space-y-2">
      <p><strong class="text-zinc-400">Why lists look empty</strong> — wrong farm selected, no data yet, telemetry not reaching the API (Pi / URL / key), or automation not configured. Compare <strong class="text-zinc-300">setpoints</strong> (targets) to <strong class="text-zinc-300">live readings</strong>.</p>
      <p>For install and logs see <strong class="text-zinc-400">docs/local-operator-bootstrap.md</strong> and <strong class="text-zinc-400">docs/operator-troubleshooting.md</strong> in the repo.</p>
      <p>For Docker/systemd capture, rotation, Loki demo stack (<strong class="text-zinc-400">docker-compose.logging.yml</strong>), and archival (vs Timescale row pruning): <strong class="text-zinc-400">docs/operator-logging-runbook.md</strong>.</p>
    </section>
  </div>
</template>

<script setup>
import HelpTip from '../components/HelpTip.vue'

const glossary = [
  {
    term: 'Zone',
    body: 'A physical grow area (room, bench, bed). Sensors and actuators attach to zones; crop cycles and many logs hang off zones.',
  },
  {
    term: 'Sensor vs live reading',
    body: 'A sensor is the logical channel; readings are time-series points from hardware or manual entry. No recent readings usually means telemetry path or device status — not “missing setpoint”.',
  },
  {
    term: 'Setpoint',
    body: 'A target band for a sensor type (min / ideal / max), often stage-scoped. Rules can compare live readings to setpoints. Different from a single raw sensor row.',
  },
  {
    term: 'Actuator & device',
    body: 'An actuator is an output (valve, pump, light). A device is the hardware that hosts actuators / bridges sensors (e.g. Pi). Automation sends commands toward actuators.',
  },
  {
    term: 'Schedule',
    body: 'Time-based automation: cron-like cadence + actions (or fertigation windows).',
  },
  {
    term: 'Rule (automation)',
    body: 'Condition-driven: when sensor predicates match, run ordered actions (cooldowns apply). Not the same as a schedule.',
  },
  {
    term: 'Task',
    body: 'Human work item — inspections, fixes, harvest prep. Often the day-to-day spine alongside automation.',
  },
  {
    term: 'Automation run',
    body: 'One execution of a schedule, rule, or program tick — success / partial / failed with details for auditing.',
  },
  {
    term: 'Farm Guardian',
    body: 'On-prem copilot chat (snapshot + optional RAG). It proposes changes like pull requests — tasks, alert ack, schedule patches, Pi pending_command — but nothing writes until you Confirm. Automation rules/alerts run separately without chat. See docs/operator-tour.md §6 and docs/farm-guardian-architecture.md §8.',
  },
  {
    term: 'Knowledge (RAG)',
    body: 'Semantic search over indexed farm text chunks from approved database domains (via rag-ingest); optional LLM answer when the API is configured. Not the same as static Help/Guide copy or Docker/API logs — see docs/rag-scope-and-threat-model.md §9.',
  },
  {
    term: 'Operational logs',
    body: 'The API prints structured lines (requests, automation outcomes, optional auth failures) to stdout/stderr — not stored in Postgres. Use LOG_FORMAT=json for stacks; retention is separate from database timeseries pruning.',
  },
]
</script>
