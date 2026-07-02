ALTER TABLE problems
    CHANGE COLUMN input_file          input_dir           VARCHAR(1024) NOT NULL,
    CHANGE COLUMN expected_output_file expected_output_dir VARCHAR(1024) NOT NULL;
