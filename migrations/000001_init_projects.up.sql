CREATE TABLE IF NOT EXISTS `languages` (
    id int AUTO_INCREMENT PRIMARY KEY,
    name varchar(255) NOT NULL,
    compile_cmd varchar(512),
    run_cmd varchar(512) NOT NULL,
    source_file varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS `problems` (
    id int AUTO_INCREMENT PRIMARY KEY,
    name varchar(255) NOT NULL,
    description text NOT NULL,
    output_format text NOT NULL,
    sample_input text NOT NULL,
    sample_output text NOT NULL,
    time_limit int NOT NULL,
    memory_limit int NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    input_file varchar(1024) NOT NULL,
    cpu_time_limit decimal,
    cpu_extra_time decimal,
    wall_time_limit decimal,
    stack_limit int,
    max_processes_and_or_threads int,
    expected_output_file varchar(1024) NOT NULL
);

CREATE TABLE IF NOT EXISTS `submissions` (
    id int AUTO_INCREMENT PRIMARY KEY,
    language_id int NOT NULL,
    problem_id int NOT NULL,
    source_code text NOT NULL,
    stdin text NOT NULL,
    stdout text NOT NULL,
    stderr text NOT NULL,
    status varchar(255) NOT NULL,
    created_at timestamp NOT NULL,
    exit_code int, 
    exit_signal int,
    finished_at timestamp,
    time_used int,
    memory_used int
    -- FOREIGN KEY (language_id) REFERENCES languages(id),
    -- FOREIGN KEY (problem_id) REFERENCES problems(id)
);