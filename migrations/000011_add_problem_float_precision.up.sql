-- Number of decimal places a floating-point answer must match. NULL means the
-- problem has no float output (compare exactly); a value N means the judge
-- accepts the user's output when every numeric field is within 10^-N of the
-- expected value (see domain.CompareOutput).
ALTER TABLE `problems`
    ADD COLUMN `float_precision` INT NULL DEFAULT NULL AFTER `max_processes_and_or_threads`;
