-- 0002 -- per-exercise sets and weight; per-session set tracking and weight used.

ALTER TABLE exercises ADD COLUMN goal_sets INTEGER;          -- default 1 if NULL at read time
ALTER TABLE exercises ADD COLUMN goal_weight_lb REAL;        -- weighted exercises

ALTER TABLE sessions ADD COLUMN set_index INTEGER;           -- 1-based; NULL for LOGGED
ALTER TABLE sessions ADD COLUMN set_total INTEGER;           -- the planned set total at log time
ALTER TABLE sessions ADD COLUMN weight_lb REAL;              -- actual weight used in this set
