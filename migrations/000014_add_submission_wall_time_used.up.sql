-- Records the peak wall-clock time (milliseconds) a submission's program took,
-- alongside the existing time_used (CPU time). Wall time is what the problem's
-- TimeLimit is now enforced against (see domain.applyExecOptions), so storing it
-- lets the API report the same quantity the judge measured. Nullable: it stays
-- NULL until the worker runs the submission and records isolate's TimeWall.
ALTER TABLE `submissions`
    ADD COLUMN `wall_time_used` INT NULL DEFAULT NULL AFTER `time_used`;
