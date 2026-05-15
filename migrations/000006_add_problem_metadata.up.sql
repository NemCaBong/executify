-- Adds richer problem metadata: clean-URL slug, difficulty bucket, public flag,
-- and denormalized counters for leaderboards / problem-list sorting.
--
-- slug and difficulty are nullable here so existing rows survive the migration;
-- backfill them, then a follow-up migration can promote slug to NOT NULL.
-- MySQL allows multiple NULLs in a UNIQUE column, so the unique index is safe.
ALTER TABLE `problems`
    ADD COLUMN `slug` VARCHAR(100) NULL AFTER `name`,
    ADD COLUMN `difficulty` ENUM('easy','medium','hard') NULL AFTER `slug`,
    ADD COLUMN `is_public` BOOLEAN NOT NULL DEFAULT TRUE AFTER `difficulty`,
    ADD COLUMN `accepted_count` INT NOT NULL DEFAULT 0 AFTER `is_public`,
    ADD COLUMN `submission_count` INT NOT NULL DEFAULT 0 AFTER `accepted_count`,
    ADD UNIQUE KEY `uk_problems_slug` (`slug`);
