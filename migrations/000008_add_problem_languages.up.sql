-- Per-language wrapper/template storage. The old single-wrapper-per-problem
-- design forced each problem to a single language; this table lets one
-- problem accept submissions in any pre-configured language by looking up
-- (problem_id, language_id) -> (template_code, wrapper_code).
CREATE TABLE IF NOT EXISTS `problem_languages` (
    problem_id    INT NOT NULL,
    language_id   INT NOT NULL,
    template_code TEXT NOT NULL,
    wrapper_code  TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (problem_id, language_id),
    INDEX idx_problem_languages_language (language_id)
);

-- Template/wrapper are now stored per (problem, language) in problem_languages.
ALTER TABLE `problems`
    DROP COLUMN `template_code`,
    DROP COLUMN `wrapper_code`;