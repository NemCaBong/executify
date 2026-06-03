-- Describes the structure of a problem's input/output files so the runtime can
-- parse each line into typed, named values (e.g. line 0 "nums" int[], line 1
-- "target" int). No foreign key to `problems` to match the existing convention
-- (see commented-out FKs in 000001_init_projects); integrity is enforced in app code.
CREATE TABLE IF NOT EXISTS `problem_io_schema` (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    problem_id INT NOT NULL,
    kind       ENUM('input', 'output') NOT NULL,
    line_index INT NOT NULL,          -- 0-based line number within the file
    key_name   VARCHAR(100) NOT NULL, -- 'nums', 'target', 'result'
    data_type  VARCHAR(50) NOT NULL,  -- 'int[]', 'int', 'string', 'int[][]'
    UNIQUE KEY uq_problem_io_schema_line (problem_id, kind, line_index),
    INDEX idx_problem_io_schema_problem (problem_id)
);