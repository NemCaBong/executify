CREATE TABLE IF NOT EXISTS `tags` (
    id int AUTO_INCREMENT PRIMARY KEY,
    name varchar(255) NOT NULL,
    slug varchar(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS `problem_tags` (
    problem_id int NOT NULL,
    tag_id int NOT NULL,
    PRIMARY KEY (problem_id, tag_id)
);
