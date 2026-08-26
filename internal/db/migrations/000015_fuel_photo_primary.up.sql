-- 000015_fuel_photo_primary.up.sql
-- is_primary was a bare boolean with nothing enforcing it. Any number of a fuel
-- entry's photos could carry it, so "the primary photo" was whichever one a
-- query happened to return first, and every consumer screen got to disagree.
--
-- Zero photos stays valid - a fuel entry without a receipt is an ordinary thing.
-- The rule is only that where photos exist, at most one is primary.

-- Existing duplicates are resolved before the index can enforce anything. The
-- lowest id wins: it is the one uploaded first, which is the closest thing to a
-- deliberate choice in data that never recorded one.
UPDATE fuel_photo SET is_primary = false
WHERE is_primary
  AND id <> (SELECT MIN(id) FROM fuel_photo p WHERE p.entry_id = fuel_photo.entry_id AND p.is_primary);

-- A partial unique index makes the invariant true regardless of code path -
-- including a direct SQL fix during an incident, which is exactly when a
-- code-level check would be bypassed.
CREATE UNIQUE INDEX uq_fuel_photo_primary ON fuel_photo (entry_id) WHERE is_primary;
