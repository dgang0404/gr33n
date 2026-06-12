package bootstraptemplates

func embeddedCatalog() *Catalog {
	templates := []Template{
		{
			TemplateKey: "jadam_indoor_photoperiod_v1", Label: "Indoor photoperiod starter (v1)",
			ShortLabel: "Indoor photoperiod v1", Tagline: "Four zones, feeding programs, inventory, and demo tasks",
			SummaryTitle: "Included in this starter pack (idempotent — the API skips duplicate rows)",
			SummaryBullets: []string{
				"Four zones: Seedling Room (indoor), Veg Room (indoor), Flower Room (indoor), Outdoor Garden (outdoor)",
				"Lighting schedules (18/6 veg, 12/12 flower) + active irrigation schedules per zone",
				"Inventory: JMS, JLF, FFJ, WCA inputs plus ready-to-use starter batches",
				"Recipes: JMS / JLF / combined drench + FFJ+WCA flowering boost with components",
				"Fertigation: 3 reservoirs, EC targets per zone, 3 programs (veg JLF, flower FFJ+WCA, outdoor JLF drench) each linked to a schedule",
				"Mixing log: 3 mixing events tied to reservoirs, programs, and inventory batches; fertigation events linked to mixes",
				"Crop cycles: active cycle per zone with primary program link",
				"Tasks: reservoir refresh tasks per zone, each linked to its irrigation schedule",
			},
			ModuleHints: []string{"zones", "fertigation", "inventory", "lighting"},
			Icon: "🌱", Recommended: true, WizardPrimary: true,
			PlaybookSection: "JADAM indoor photoperiod (`jadam_indoor_photoperiod_v1`)",
			RelatedCommonsSlug: "gr33n-cultivator-seed-pack-v1", SortOrder: 10,
		},
		{
			TemplateKey: "greenhouse_climate_v1", Label: "Greenhouse climate (v1)",
			ShortLabel: "Greenhouse v1", Tagline: "Shade, vents, humidity bands, and Pi placeholder",
			SummaryTitle: "Greenhouse / tent climate (dew point, VPD, CO2 — pair with Pi derived sensors)",
			SummaryBullets: []string{
				"One zone: Greenhouse + Pi device placeholder",
				"Sensors: air temp, RH, CO2, dew point, VPD (Pa)",
				"Actuators: exhaust fan, humidifier, dehumidifier, shade motor, CO2 injector",
				"Automation rules (inactive): dew/VPD/CO2/temperature thresholds → equipment",
				"Task: weekly CO2 / enrichment checklist",
			},
			ModuleHints: []string{"zones", "greenhouse", "climate"},
			Icon: "🏠", WizardPrimary: true, SortOrder: 20,
			PlaybookSection: "Greenhouse climate (`greenhouse_climate_v1`)",
		},
		{
			TemplateKey: "chicken_coop_v1", Label: "Chicken coop (v1)", ShortLabel: "Chicken coop v1",
			Tagline: "Coop sensors, feeder, and climate actuators",
			SummaryTitle: "Chicken coop starter (sensors, actuators, schedules, rules — tune before enabling rules)",
			SummaryBullets: []string{
				"One zone: Chicken Coop + Pi device placeholder",
				"Sensors: water level, feed level, air temperature, humidity",
				"Actuators: feeder hopper, water valve, exhaust fan, heat lamp",
				"Schedules: morning / evening feed reminders (inactive by default)",
				"Automation rules (inactive): low water / low feed → tasks; hot → fan; cold → heat lamp",
				"Task: weekly egg collection reminder",
			},
			ModuleHints: []string{"zones", "animals", "climate"}, SortOrder: 30,
			PlaybookSection: "Chicken coop (`chicken_coop_v1`)",
		},
		{
			TemplateKey: "drying_room_v1", Label: "Drying / cure room (v1)", ShortLabel: "Drying room v1",
			Tagline: "Post-harvest environment monitoring",
			SummaryTitle: "Drying / cure room (defaults skew cannabis; retune for basil, orchids, herbs)",
			SummaryBullets: []string{
				"One zone: Drying Room + Pi device placeholder",
				"Sensors: temperature, humidity, dew point",
				"Actuators: dehumidifier, circulation fan",
				"Automation rules (inactive): dew-point on/off band + high-RH circulation",
				"Task: daily environment log reminder",
			},
			ModuleHints: []string{"zones", "climate", "harvest"}, SortOrder: 40,
			PlaybookSection: "Drying room (`drying_room_v1`)",
		},
		{
			TemplateKey: "small_aquaponics_v1", Label: "Small aquaponics (v1)", ShortLabel: "Aquaponics v1",
			Tagline: "Fish tank + grow bed loop starter",
			SummaryTitle: "Small aquaponics loop (fish tank + grow bed)",
			SummaryBullets: []string{
				"Two zones: Fish Tank, Grow Bed + Pi device placeholder",
				"Tank sensors: water temperature, pH, ammonia, nitrate; bed sensors: pH, EC",
				"Actuators: return pump, air pump",
				"gr33naquaponics.loops row: Main aquaponics loop (meta documents zone names)",
				"Schedule: daily fish-feed reminder (inactive)",
				"Automation rules (inactive): ammonia spike → task; cold tank → task",
				"Task: daily feed fish reminder",
			},
			ModuleHints: []string{"zones", "aquaponics", "water"}, SortOrder: 50,
			PlaybookSection: "Small aquaponics (`small_aquaponics_v1`)",
		},
	}
	cat := &Catalog{byKey: make(map[string]Template, len(templates))}
	for _, t := range templates {
		cat.byKey[t.TemplateKey] = t
		cat.list = append(cat.list, t)
	}
	return cat
}
