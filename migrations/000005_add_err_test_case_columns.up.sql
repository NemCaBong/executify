ALTER TABLE submissions
    ADD COLUMN err_test_case_input  TEXT NOT NULL DEFAULT '',
    ADD COLUMN err_test_case_output TEXT NOT NULL DEFAULT '';