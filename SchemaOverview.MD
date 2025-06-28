## gr33n Project & Schema Overview (Updated)

**Project Name:** gr33n

**Overall Goal:**
The primary goal of the **gr33n** project is to develop a comprehensive, highly adaptable, and scalable open-source agricultural management platform. It aims to empower farmers by providing a suite of tools to optimize operations, improve efficiency, support data-driven decision-making, and promote sustainable farming practices . The platform is designed to cater to a wide spectrum of users, from small-scale homesteads ("little system") to large, complex industrial farming operations ("big system"), with the capability to be deployed online, offline, or on an intranet.

**Core Objectives & Design Philosophy:**

1.  **Modularity:** The system is built upon a modular architecture, where distinct agricultural domains (e.g., core farm management, crop cultivation, animal husbandry, inventory, equipment, irrigation, aquaponics, natural farming inputs) are encapsulated within their own well-defined schemas within a PostgreSQL database. This approach aligns with principles seen in generic and modular farm model designs .
2.  **Adaptability & Scalability:** The key to serving both "big" and "little" systems lies in a central configuration mechanism (`gr33n_core.farm_active_modules`). This allows each farm instance to:
    *   Enable only the specific functional modules they require.
    *   Configure the complexity level within active modules (e.g., simple vs. advanced inventory tracking).
    This effectively "prunes" the system's presented complexity for smaller users while offering the full suite of features for larger operations, all on a unified underlying schema. The `gr33n_core.farms.scale_tier` field further aids the application layer in tailoring the user experience.
3.  **Comprehensive Data Management:** The schema provides detailed structures for:
    *   **Core Farm Setup:** Users, farm details, multi-user memberships, hierarchical zones.
    *   **IoT Integration:** Devices, sensors (with time-series readings via `gr33n_core.sensor_readings`), and actuators (with event logging via `gr33n_core.actuator_events`). Actuators are generic to control elements like irrigation valves, animal feed/water dispensers, and gates.
    *   **Automation:** Robust scheduling (`gr33n_core.schedules`), complex "if-this-then-that" automation rules (`gr33n_core.automation_rules`), and a unified system for defining actions (`gr33n_core.executable_actions`).
    *   **Communication:** Templated notifications and an alert logging system (`gr33n_core.notification_templates`, `gr33n_core.alerts_notifications`).
    *   **Specialized Agricultural Modules:** Detailed tables within schemas like `gr33n_crops`, `gr33n_animals`, `gr33n_inventory`, `gr33n_equipment` (including field operations and maintenance), `gr33n_irrigation`, `gr33n_aquaponics`, and `gr33n_natural_farming` to capture the nuances of diverse farming practices.
4.  **Interoperability & Extensibility:** While specific API design is the next step, the schema is structured to support well-defined data exchange. The use of `JSONB` for configurations and metadata allows for flexibility and future expansion without immediate schema alterations. The open-source nature can foster community development and integration, similar to initiatives like OpenAgri .
5.  **Ease of Use & Flexible Setup:** A key objective is to make the platform accessible.
    *   **Comprehensive Documentation (Wiki):** A detailed wiki will provide step-by-step instructions for manual setup and configuration, enabling users to deploy **gr33n** offline or on a private intranet without internet dependency. This documentation will guide users in tailoring active modules and their complexity.
    *   **Optional AI-Assisted Setup:** For users with internet connectivity, an AI-driven setup assistant could optionally guide them through a questionnaire to automatically configure active modules and complexity levels, tailoring **gr33n** to their specific needs.

**Technological Foundation (Schema):**
The schema is designed for PostgreSQL and is compatible with cloud database services like Supabase (leveraging its `auth.users` table if used in an online mode) or a self-hosted PostgreSQL instance for offline/intranet deployments. It strongly recommends the use of the TimescaleDB extension for managing large volumes of time-series data (sensor readings, actuator events, logs) efficiently, and PostGIS for advanced geospatial capabilities if required. Row Level Security (RLS) is intended to be implemented comprehensively to ensure data isolation and security between different farm tenants.

**Overall, the gr33n schema provides a robust, detailed, and flexible relational database blueprint designed to be the backbone of a modern, data-driven, and highly customizable agricultural management platform that can scale with the diverse needs of its users and operate in various deployment environments.**

---

This updated overview statement now clearly incorporates your preference for a wiki-based setup as the primary documented method, with the AI chat box as an optional convenience for online scenarios. This reinforces the offline-first capability you're aiming for.
