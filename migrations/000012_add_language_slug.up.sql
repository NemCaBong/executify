-- A URL-safe identifier for a language, used to fetch a problem's template
-- code for a specific language (see GET /problems/:slug/languages/:lang).
-- Backfilled from name (lowercased, spaces -> hyphens); adjust per language
-- as needed. Unique so a slug maps to exactly one language.
ALTER TABLE `languages`
    ADD COLUMN `slug` varchar(255) NOT NULL DEFAULT '' AFTER `name`;

UPDATE `languages`
    SET `slug` = LOWER(REPLACE(`name`, ' ', '-'))
    WHERE `slug` = '';

CREATE UNIQUE INDEX `idx_languages_slug` ON `languages` (`slug`);
