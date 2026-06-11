/**
 * Derives 2–3 contextual follow-up prompt chips from a completed Guardian turn.
 * These appear below the last assistant response so the user can continue the
 * conversation with one tap — the "chat loop" pattern.
 *
 * @param {string} userMessage   - the message the user just sent
 * @param {string} assistantMessage - Guardian's response
 * @returns {Array<{id: string, label: string, message: string}>}
 */
export function deriveFollowUps(userMessage, assistantMessage) {
  const text = `${userMessage}\n${assistantMessage}`.toLowerCase()
  const userText = userMessage.toLowerCase()

  const candidates = []

  // Specific plant topics come first so their chips aren't crowded out by
  // generic lighting/watering matches that appear in almost every response.

  // ── Cannabis / photoperiod ────────────────────────────────────────────────
  if (/cannabis|marijuana|weed|hemp|flower\s*stage|12\s*\/?\s*12|18\s*\/?\s*6|photoperiod/.test(text)) {
    if (!/how.*(flip|switch|transition)|flip.*(12|flower)/.test(userText)) {
      candidates.push({
        id: 'flip-schedule',
        label: 'When should I flip to 12/12?',
        message: 'When should I flip my cannabis to 12/12 for flowering, and what signs tell me the plant is ready?',
      })
    }
    if (!/harvest|trichome|ready to harvest/.test(userText)) {
      candidates.push({
        id: 'harvest-window',
        label: 'How do I know when to harvest?',
        message: 'How do I tell when cannabis is ready to harvest — what should I look for in the trichomes and pistils?',
      })
    }
    candidates.push({
      id: 'cannabis-vpd',
      label: 'What VPD should I target?',
      message: 'What VPD should I target for cannabis at each stage — seedling, veg, and flower?',
    })
  }

  // ── Orchid ────────────────────────────────────────────────────────────────
  if (/orchid|phalaenopsis|dendrobium|cattleya|vanda/.test(text)) {
    if (!/repot/.test(userText)) {
      candidates.push({
        id: 'orchid-repot',
        label: 'When should I repot?',
        message: 'When and how should I repot my orchid — what substrate works best and how do I handle aerial roots?',
      })
    }
    candidates.push({
      id: 'orchid-rebloom',
      label: 'How do I trigger reblooming?',
      message: 'How do I get my orchid to rebloom — what temperature drop or light change encourages a new spike?',
    })
  }

  // ── Eggplant / fruiting vegetables ───────────────────────────────────────
  if (/eggplant|aubergine|tomato|pepper|capsicum|solanum melongena/.test(text)) {
    if (!/pollinat/.test(userText)) {
      candidates.push({
        id: 'pollination',
        label: 'How do I hand-pollinate indoors?',
        message: 'How do I hand-pollinate indoor eggplant or tomatoes to get fruit set without bees?',
      })
    }
    candidates.push({
      id: 'fruiting-nutrients',
      label: 'What nutrients boost fruiting?',
      message: 'What nutrient ratio shift does my plant need once it starts fruiting — how much more P and K vs veg?',
    })
  }

  // ── Ramps / wild alliums ──────────────────────────────────────────────────
  if (/\bramp(s)?\b|allium tricoccum|wild garlic|wild leek/.test(text)) {
    candidates.push({
      id: 'ramps-dormancy',
      label: 'When do ramps go dormant?',
      message: 'When do ramps go dormant and how should I care for the bulbs through summer?',
    })
    if (!/harvest ramp/.test(userText)) {
      candidates.push({
        id: 'ramps-harvest',
        label: 'How do I harvest ramps sustainably?',
        message: 'How do I harvest ramps without destroying the colony — leaves only, or can I take bulbs too, and when?',
      })
    }
  }

  // ── EC / nutrients ────────────────────────────────────────────────────────
  if (/\bec\b|electrical conductivity|nutrient|feeding|fertigation|ppm\b|tds\b/.test(text)) {
    if (!/measure ec|check ec|ec meter/.test(userText)) {
      candidates.push({
        id: 'ec-measure',
        label: 'How do I measure EC?',
        message: 'How do I measure EC correctly — what tool do I need, and when in the feeding cycle should I check it?',
      })
    }
    candidates.push({
      id: 'ec-runoff',
      label: 'What runoff EC is okay?',
      message: 'What runoff EC reading should I expect after feeding, and when does it signal a buildup problem I need to flush?',
    })
  }

  // ── pH ────────────────────────────────────────────────────────────────────
  if (/\bph\b|acid|alkalin|ph drift/.test(text)) {
    if (!/fix ph|adjust ph|correct ph/.test(userText)) {
      candidates.push({
        id: 'ph-adjust',
        label: 'How do I fix pH drift?',
        message: 'How do I correct pH drift in my reservoir — what products do I use and how often should I check?',
      })
    }
    candidates.push({
      id: 'ph-symptoms',
      label: 'What do pH problems look like?',
      message: 'What do pH problems look like on leaves — how do I tell a pH lockout from a nutrient deficiency?',
    })
  }

  // ── Alerts / issues ───────────────────────────────────────────────────────
  if (/\balert|alarm|warn|offline|error|deficien|symptom|yellowing|browning/.test(text)) {
    candidates.push({
      id: 'alert-next-step',
      label: "What's my most urgent issue?",
      message: 'What is the most urgent issue I should address right now based on what you just described?',
    })
  }

  // ── Program / cycle / schedule ───────────────────────────────────────────
  if (/\bprogram\b|cycle\b|schedule\b|fertigation|run.*feed|feed.*run/.test(text)) {
    if (!/next.*run|when.*run/.test(userText)) {
      candidates.push({
        id: 'next-program-run',
        label: 'When does my program run next?',
        message: 'When does my current feeding program run next, and is there anything I should check before it fires?',
      })
    }
  }

  // ── Watering / irrigation ─────────────────────────────────────────────────
  if (/\bwater(ing)?\b|irrigation|drip|runoff|moisture|overwater/.test(text)) {
    if (!/when to water|signs.*water/.test(userText)) {
      candidates.push({
        id: 'watering-signs',
        label: 'How do I know when to water?',
        message: 'What are the signs my plants need watering vs overwatering — how do I read the plant and the substrate?',
      })
    }
    if (/drip|fertigation|program|schedule|cycle/.test(text)) {
      candidates.push({
        id: 'drip-frequency',
        label: 'How often should I run the drip?',
        message: 'How often should I run drip irrigation for my current crop and substrate, and how do I adjust based on plant response?',
      })
    }
  }

  // ── Lighting / DLI ────────────────────────────────────────────────────────
  if (/\blight(ing)?\b|dli\b|ppfd|lux\b|spectrum|led\b|hps\b|photoperiod/.test(text)) {
    if (!/dli target|what dli/.test(userText)) {
      candidates.push({
        id: 'dli-target',
        label: 'What DLI does my crop need?',
        message: 'What daily light integral (DLI) should I target for my crop at this stage, and how do I calculate it from my fixture specs?',
      })
    }
    if (!/dim|reduce|lower.*light/.test(userText)) {
      candidates.push({
        id: 'light-intensity',
        label: 'Should I adjust my light intensity?',
        message: 'Should I dim or adjust my light intensity right now based on the current crop stage and VPD readings?',
      })
    }
  }

  // ── VPD / climate ─────────────────────────────────────────────────────────
  if (/\bvpd\b|vapor pressure|temp(erature)?|humidity|\brh\b|climate|comfort/.test(text)) {
    if (!/vpd range|what vpd/.test(userText)) {
      candidates.push({
        id: 'vpd-bands',
        label: 'What VPD range is ideal?',
        message: 'What VPD range is ideal for my current grow stage, and how should I balance temperature and humidity to hit it?',
      })
    }
  }

  // ── Generic fallback when nothing specific matched ─────────────────────────
  if (candidates.length === 0) {
    candidates.push(
      {
        id: 'grow-status',
        label: 'What needs attention now?',
        message: 'Give me a plain-language summary of what needs attention in my grow right now — alerts, feeds, or climate.',
      },
      {
        id: 'next-action',
        label: "What's my next action?",
        message: 'What is the single most important action I should take today to keep my grow on track?',
      },
      {
        id: 'optimize-yield',
        label: 'How do I improve yield?',
        message: 'What changes to my environment or feeding would most improve my yield from here?',
      },
    )
  }

  // Dedupe and cap at 3
  const seen = new Set()
  return candidates
    .filter((c) => {
      if (seen.has(c.id)) return false
      seen.add(c.id)
      return true
    })
    .slice(0, 3)
}
