-- migrate_facilities_from_csv.sql
-- Loads/updates facilities from facilities_private.csv into the facilities table.
-- CSV columns (header required):
-- uid,name,subcounty_uid,subcounty,district_uid,district,region_uid,region,hflevel,ownership

BEGIN;

-- 1) Staging table matching CSV exactly
DROP TABLE IF EXISTS facilities_csv_staging;

CREATE TABLE facilities_csv_staging (
    uid           TEXT,
    name          TEXT,
    subcounty_uid TEXT,
    subcounty     TEXT,
    district_uid  TEXT,
    district      TEXT,
    region_uid    TEXT,
    region        TEXT,
    hflevel       TEXT,
    ownership     TEXT
);

-- 2) Load CSV into staging (adjust path if needed)
\copy facilities_csv_staging FROM 'facilities_private.csv' WITH (FORMAT csv, HEADER true);

-- 3) Upsert into main facilities table by uid
-- Mapping:
--   uid       -> facilities.uid
--   name      -> facilities.name
--   subcounty -> facilities.subcounty
--   district  -> facilities.district
--   hflevel   -> facilities.level
--   code      -> uid
--   client_code_prefix -> first 5 chars of uid (or uid if shorter)
INSERT INTO facilities (
    id,
    created_at,
    updated_at,
    deleted_at,
    uid,
    name,
    code,
    level,
    subcounty,
    hsd,
    district,
    client_code_prefix
)
SELECT
    gen_random_uuid() AS id,
    NOW()             AS created_at,
    NOW()             AS updated_at,
    NULL::timestamptz AS deleted_at,
    TRIM(uid)         AS uid,
    TRIM(name)        AS name,
    TRIM(uid)         AS code,
    TRIM(hflevel)     AS level,
    TRIM(subcounty)   AS subcounty,
    ''::text          AS hsd,
    TRIM(district)    AS district,
    UPPER(
      TRIM(
        CASE
          WHEN uid IS NULL OR uid = '' THEN ''
          WHEN length(uid) <= 5        THEN uid
          ELSE substring(uid FROM 1 FOR 5)
        END
      )
    )                 AS client_code_prefix
FROM facilities_csv_staging
WHERE TRIM(uid)  <> ''
  AND TRIM(name) <> ''
ON CONFLICT (uid) DO UPDATE
SET
    name               = EXCLUDED.name,
    code               = EXCLUDED.code,
    level              = EXCLUDED.level,
    subcounty          = EXCLUDED.subcounty,
    hsd                = EXCLUDED.hsd,
    district           = EXCLUDED.district,
    client_code_prefix = EXCLUDED.client_code_prefix,
    updated_at         = NOW()
WHERE facilities.deleted_at IS NULL
  AND (
      facilities.name               IS DISTINCT FROM EXCLUDED.name OR
      facilities.code               IS DISTINCT FROM EXCLUDED.code OR
      facilities.level              IS DISTINCT FROM EXCLUDED.level OR
      facilities.subcounty          IS DISTINCT FROM EXCLUDED.subcounty OR
      facilities.hsd                IS DISTINCT FROM EXCLUDED.hsd OR
      facilities.district           IS DISTINCT FROM EXCLUDED.district OR
      facilities.client_code_prefix IS DISTINCT FROM EXCLUDED.client_code_prefix
  );

-- 4) Clean up staging
DROP TABLE facilities_csv_staging;

COMMIT;
