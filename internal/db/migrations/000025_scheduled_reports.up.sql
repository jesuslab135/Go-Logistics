-- 000025_scheduled_reports.up.sql
-- Reports the system builds on a clock and emails to the tenant company:
-- a weekly fuel report and a monthly maintenance cost report.
--
-- Until now nothing in this deployment ran without a request (see the note on
-- PruneNotifications). The API process now carries a ticker, and these two
-- tables are what make it safe: report_recipient says who is told, and
-- report_run remembers what was already sent.

CREATE TABLE report_recipient (
    id          bigserial    PRIMARY KEY,
    company_id  bigint       NOT NULL,
    report_kind varchar(30)  NOT NULL,
    email       varchar(254) NOT NULL,
    is_active   boolean      NOT NULL DEFAULT true,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT ck_report_recipient_kind
        CHECK (report_kind IN ('fuel_weekly', 'maintenance_monthly'))
);

ALTER TABLE report_recipient
    ADD CONSTRAINT fk_report_recipient_company
    FOREIGN KEY (company_id) REFERENCES company (id) ON DELETE CASCADE;

-- Case-insensitive: Flota@Empresa.mx and flota@empresa.mx are one mailbox, and
-- listing it twice would deliver the same report twice.
CREATE UNIQUE INDEX uq_report_recipient_company_kind_email
    ON report_recipient (company_id, report_kind, lower(email));

-- One row per company, report and period. The unique constraint is the
-- send-once guarantee: whoever inserts the row owns the run, and a second API
-- container, a restart or a manual CLI run finds it already there.
--
-- period_start is the first local day of the period, in the company's own
-- timezone. It is a date, not a timestamp, so that the same week is the same
-- row regardless of the offset it was computed in.
CREATE TABLE report_run (
    id          bigserial   PRIMARY KEY,
    company_id  bigint      NOT NULL,
    report_kind varchar(30) NOT NULL,
    period_start date       NOT NULL,
    status      varchar(30) NOT NULL,
    attempts    integer     NOT NULL DEFAULT 0,
    error       text,
    recipients  integer     NOT NULL DEFAULT 0,
    started_at  timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    CONSTRAINT ck_report_run_kind
        CHECK (report_kind IN ('fuel_weekly', 'maintenance_monthly')),
    CONSTRAINT ck_report_run_status
        CHECK (status IN ('running', 'sent', 'not_sent', 'skipped_empty', 'skipped_no_recipients', 'failed')),
    CONSTRAINT uq_report_run_company_kind_period
        UNIQUE (company_id, report_kind, period_start)
);

ALTER TABLE report_run
    ADD CONSTRAINT fk_report_run_company
    FOREIGN KEY (company_id) REFERENCES company (id) ON DELETE CASCADE;

CREATE INDEX idx_report_run_company_started_at
    ON report_run (company_id, started_at DESC);

-- fuel_entry had only its primary key. The weekly report reads a date range
-- per asset, which without this is a scan of the company's whole fuel history.
CREATE INDEX idx_fuel_entry_asset_date ON fuel_entry (asset_id, date);
