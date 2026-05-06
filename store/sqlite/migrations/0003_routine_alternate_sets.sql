-- 0003 -- routines can run in alternate-sets (circuit) mode.

ALTER TABLE routines ADD COLUMN alternate_sets INTEGER NOT NULL DEFAULT 0;
