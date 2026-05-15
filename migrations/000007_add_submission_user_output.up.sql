-- Adds user_output to capture the user program's stdout (regular print()
-- statements), which is separate from actual_output (the answer written to
-- fd 3 by the wrapper). Useful for showing users their debug prints alongside
-- the verdict.
ALTER TABLE `submissions`
    ADD COLUMN `user_output` TEXT NOT NULL DEFAULT '' AFTER `actual_output`;