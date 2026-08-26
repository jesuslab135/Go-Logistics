-- 000016_trailer_classification.up.sql
-- classification and classification_2 were varchar(100) free text - in Django
-- too, where the sibling fields trailer_type and financing both carry choices.
-- So every fleet spelled its own vocabulary: "Dry Van", "DRY VAN" and "dry van"
-- are three classifications as far as any filter or report is concerned.
--
-- A company-scoped catalog rather than an enum. The values are ops vocabulary
-- nobody has written down and they differ per fleet, and every comparable list
-- in this system - fuel_type, measurement_unit, asset_status, issue_priority -
-- is already a company-scoped catalog. Following that is better than inventing
-- a second pattern for the same kind of thing.
CREATE TABLE trailer_classification (
    id         bigserial    PRIMARY KEY,
    company_id bigint       NOT NULL,
    name       varchar(100) NOT NULL,
    position   integer      NOT NULL DEFAULT 0,
    CONSTRAINT uq_trailer_classification UNIQUE (company_id, name)
);

ALTER TABLE trailer_classification ADD CONSTRAINT fk_trailer_classification_company
    FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE;

-- Seeded from each company's own data, not from a guessed list. Seeding a
-- vocabulary somebody else invented would put words in ops' mouths; every value
-- here is one that fleet already uses.
INSERT INTO trailer_classification (company_id, name)
SELECT DISTINCT a.company_id, v
FROM trailer t
JOIN asset a ON a.id = t.asset_id
CROSS JOIN LATERAL (VALUES (btrim(t.classification)), (btrim(t.classification_2))) AS c(v)
WHERE btrim(v) <> ''
ON CONFLICT DO NOTHING;

ALTER TABLE trailer
    ADD COLUMN classification_id   bigint,
    ADD COLUMN classification_2_id bigint;

ALTER TABLE trailer ADD CONSTRAINT fk_trailer_classification
    FOREIGN KEY (classification_id) REFERENCES trailer_classification(id) ON DELETE SET NULL;
ALTER TABLE trailer ADD CONSTRAINT fk_trailer_classification_2
    FOREIGN KEY (classification_2_id) REFERENCES trailer_classification(id) ON DELETE SET NULL;

UPDATE trailer t SET classification_id = tc.id
FROM asset a, trailer_classification tc
WHERE a.id = t.asset_id AND tc.company_id = a.company_id AND tc.name = btrim(t.classification);

UPDATE trailer t SET classification_2_id = tc.id
FROM asset a, trailer_classification tc
WHERE a.id = t.asset_id AND tc.company_id = a.company_id AND tc.name = btrim(t.classification_2);

-- The varchar columns stay for one release, per the additive-migration policy:
-- a deployed client still reads them, and dropping a column something reads is
-- an outage where keeping a dead one for a release costs nothing.
COMMENT ON COLUMN trailer.classification IS
    'Deprecated: superseded by classification_id. Retained for one release; no longer written.';
COMMENT ON COLUMN trailer.classification_2 IS
    'Deprecated: superseded by classification_2_id. Retained for one release; no longer written.';
