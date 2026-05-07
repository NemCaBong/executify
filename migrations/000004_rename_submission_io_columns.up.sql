ALTER TABLE submissions
  CHANGE COLUMN stdin   `input`         text NOT NULL,
  CHANGE COLUMN stdout  actual_output   text NOT NULL,
  ADD COLUMN    expected_output         text NOT NULL DEFAULT '';