-- External-facing identifier for users. UUIDv7 (time-ordered, sortable) is
-- generated in the application layer via a GORM BeforeCreate hook — see
-- internal/adapter/repository/entity/user.go. The DB-level DEFAULT (UUID())
-- is a v4 fallback so direct INSERTs (e.g. manual rows) still get a value;
-- the v4 fallback also backfills existing rows during this ALTER.
ALTER TABLE `users`
    ADD COLUMN `uuid` CHAR(36) NOT NULL DEFAULT (UUID()) AFTER `id`,
    ADD UNIQUE KEY `uk_users_uuid` (`uuid`);
