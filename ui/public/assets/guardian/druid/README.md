# Guardian druid state artwork (hand-drawn only)

Human artists drop **hand-drawn** illustrations here — one image per Guardian awakening state.
**Do not use AI-generated art** for final production assets.

**Shipped placeholders:** six minimal SVG silhouettes (`*.svg`) with a faint `PLACEHOLDER`
watermark — swap them out by dropping real art and updating `manifest.json` (set
`"placeholder": false` or remove that key when done).

## States

| File key (`manifest.json`) | When it shows |
|----------------------------|---------------|
| `sleeping` | Guardian is cold; login may trigger background warmup |
| `dormant` | Operator chose **Rest now** or auto-rest after idle |
| `stirring` | Model warmup / awakening in progress |
| `ready` | Counsel model loaded; chat is available |
| `busy` | Guardian is answering a question |
| `unavailable` | Ollama down, misconfigured, or AI disabled |

## Deliverables

1. **Image file** in this folder — preferred **WebP** or **PNG**, square **256×256** (512×512 @2x optional).
2. **Manifest entry** — register the filename so the UI loads it.

Example after adding resting art:

```json
{
  "version": 1,
  "files": {
    "dormant": "dormant.webp"
  }
}
```

## Tone

Warm druid caretaker — not cosplay, not mascot overload. Match copy in `GuardianAwakeningPanel` (“The Guardian rests…”, “stirring…”). Subtle farm/green palette fits the app; transparent background recommended.

## Wiring

- `ui/src/lib/guardianStateArt.js` — manifest + URL helpers
- `ui/src/components/GuardianStateArt.vue` — image slot (hidden until manifest + file exist)
- Shown in **Settings → Farm Guardian readiness** and the chat **awakening panel**

Until `files` is non-empty, the UI keeps today’s text badges only. Placeholder SVGs are
registered by default — replace one state at a time when hand-drawn art is ready.
