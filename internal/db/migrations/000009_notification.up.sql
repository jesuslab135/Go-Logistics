-- BR-15: notification persistence. Django had no equivalent — the endpoint was
-- a stub that always returned an empty list — so this table is designed here
-- rather than ported.
--
-- A notification belongs to one employee within one company: the same person
-- working for two companies must not see one company's notifications while
-- acting for the other.
CREATE TABLE notification (
    id          bigserial PRIMARY KEY,
    company_id  bigint       NOT NULL,
    employee_id bigint       NOT NULL,
    kind        varchar(50)  NOT NULL,
    title       varchar(200) NOT NULL,
    body        text         NOT NULL DEFAULT '',
    url         text         NOT NULL DEFAULT '',
    read_at     timestamptz,
    created_at  timestamptz  NOT NULL
);

ALTER TABLE notification ADD CONSTRAINT fk_notification_company FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE;
ALTER TABLE notification ADD CONSTRAINT fk_notification_employee FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE;

-- Serves both the inbox page and the unread count, which share this predicate.
CREATE INDEX idx_notification_recipient ON notification (employee_id, company_id, created_at DESC);
