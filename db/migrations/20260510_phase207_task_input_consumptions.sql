-- Phase 20.7 WS1 — task_input_consumptions.
-- Most of the Phase 20.7 WS1 scope was pre-shipped by Phase 20.95 WS2:
-- * input_definitions.unit_cost / unit_cost_currency / unit_cost_unit_id
-- * input_batches.low_stock_threshold
-- * actuators.watts
-- * gr33ncore.farm_energy_prices table (+ CRUD routes)
-- * cost_transactions.crop_cycle_id
-- * input_category_enum broadened with animal_feed / bedding / veterinary_supply
--
-- All that's left for 20.7 WS1 is the manual-consumption join table: a task
-- (e.g. "top-dress row 3 with FAA") records what it drew from inventory, and
-- the autologger decrements the batch + writes a cost_transaction just like
-- the mixing-event path. DELETE reverses the pair (credit batch, write a
-- compensating cost_transaction so the ledger stays append-only).
--
-- Additive: ON CONFLICT-safe, nothing existing changes shape.

CREATE TABLE IF NOT EXISTS gr33ncore.task_input_consumptions (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    task_id         BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
    input_batch_id  BIGINT NOT NULL REFERENCES gr33nnaturalfarming.input_batches(id) ON DELETE RESTRICT,
    quantity        NUMERIC(10,3) NOT NULL CHECK (quantity > 0),
    unit_id         BIGINT NOT NULL REFERENCES gr33ncore.units(id) ON DELETE RESTRICT,
    notes           TEXT,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    recorded_by     UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    -- The autologger stamps this on insert; DELETE uses it to find and
    -- void the paired cost_transactions row. NULL means the batch had no
    -- unit_cost at consumption time (stock was still decremented).
    cost_transaction_id BIGINT REFERENCES gr33ncore.cost_transactions(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_task_consumptions_task
    ON gr33ncore.task_input_consumptions (task_id);
CREATE INDEX IF NOT EXISTS idx_task_consumptions_batch
    ON gr33ncore.task_input_consumptions (input_batch_id);
CREATE INDEX IF NOT EXISTS idx_task_consumptions_farm
    ON gr33ncore.task_input_consumptions (farm_id);
DROP TRIGGER IF EXISTS trg_task_input_consumptions_updated_at
    ON gr33ncore.task_input_consumptions;
CREATE TRIGGER trg_task_input_consumptions_updated_at
    BEFORE UPDATE ON gr33ncore.task_input_consumptions
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
