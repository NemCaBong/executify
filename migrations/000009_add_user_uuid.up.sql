-- External-facing identifier for users. UUIDv7 (time-ordered, sortable) is
-- generated in the application layer via a GORM BeforeCreate hook — see
-- internal/adapter/repository/entity/user.go. The DB-level DEFAULT (UUID())
-- is a v4 fallback so direct INSERTs (e.g. manual rows) still get a value.
--
-- The steps are split deliberately: a single `ADD COLUMN ... DEFAULT (UUID())`
-- embeds a non-deterministic function in the DDL that MySQL writes to the
-- binlog, which fails with Error 1674 (unsafe system function). Adding the
-- column first and declaring the functional default via a separate
-- `ALTER COLUMN ... SET DEFAULT` keeps the default a pure metadata change that
-- is never evaluated against existing rows, so it is binlog-safe.
ALTER TABLE `users`
    ADD COLUMN `uuid` CHAR(36) NOT NULL AFTER `id`;

ALTER TABLE `users`
    ALTER COLUMN `uuid` SET DEFAULT (UUID());

ALTER TABLE `users`
    ADD UNIQUE KEY `uk_users_uuid` (`uuid`);
