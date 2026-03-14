-- HMIS MCH 007: Integrated Family Planning Register – Initial Schema
-- This migration is provided as reference; GORM AutoMigrate handles table creation.
-- Run this manually if you prefer SQL-driven migrations.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Facilities
CREATE TABLE IF NOT EXISTS facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    code VARCHAR(20) NOT NULL,
    level VARCHAR(20),
    subcounty VARCHAR(100),
    hsd VARCHAR(100),
    district VARCHAR(100),
    client_code_prefix VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_facilities_code ON facilities(code) WHERE deleted_at IS NULL;

-- Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(200) NOT NULL,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'facility_user',
    facility_id UUID REFERENCES facilities(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_token ON refresh_tokens(token) WHERE deleted_at IS NULL;

-- Option sets (lookup tables for all coded values)
CREATE TABLE IF NOT EXISTS option_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(50) NOT NULL,
    code VARCHAR(20) NOT NULL,
    label VARCHAR(200) NOT NULL,
    description VARCHAR(500),
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_option_set_cat_code ON option_sets(category, code) WHERE deleted_at IS NULL;

-- Client number sequence (per facility per day)
CREATE TABLE IF NOT EXISTS client_number_seqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id UUID NOT NULL REFERENCES facilities(id),
    seq_date DATE NOT NULL,
    last_seq INT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_client_number_seq_unique ON client_number_seqs(facility_id, seq_date);

-- FP Registrations (main register entries)
CREATE TABLE IF NOT EXISTS fp_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id UUID NOT NULL REFERENCES facilities(id),
    created_by UUID NOT NULL REFERENCES users(id),

    visit_date VARCHAR(10) NOT NULL,
    serial_number INT NOT NULL,
    client_number VARCHAR(30),
    is_visitor BOOLEAN DEFAULT FALSE,

    nin VARCHAR(30),
    surname VARCHAR(100) NOT NULL,
    given_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),

    village VARCHAR(100),
    parish VARCHAR(100),
    subcounty VARCHAR(100),
    district VARCHAR(100),

    sex VARCHAR(1) NOT NULL,
    age INT NOT NULL,

    is_new_user BOOLEAN DEFAULT FALSE,
    is_revisit BOOLEAN DEFAULT FALSE,
    previous_method VARCHAR(50),

    hts_code VARCHAR(10),

    counseling_individual BOOLEAN DEFAULT FALSE,
    counseling_as_couple BOOLEAN DEFAULT FALSE,
    counseling_om BOOLEAN DEFAULT FALSE,
    counseling_se BOOLEAN DEFAULT FALSE,
    counseling_wd BOOLEAN DEFAULT FALSE,
    counseling_ms BOOLEAN DEFAULT FALSE,

    is_switching BOOLEAN DEFAULT FALSE,
    switching_reason VARCHAR(10),

    pills_coc_cycles INT DEFAULT 0,
    pills_pop_cycles INT DEFAULT 0,
    pills_ecp_pieces INT DEFAULT 0,

    condoms_male_units INT DEFAULT 0,
    condoms_female_units INT DEFAULT 0,

    injectable_dmpa_im_doses INT DEFAULT 0,
    injectable_dmpa_sc_pa_doses INT DEFAULT 0,
    injectable_dmpa_sc_si_doses INT DEFAULT 0,

    implant_3_years BOOLEAN DEFAULT FALSE,
    implant_5_years BOOLEAN DEFAULT FALSE,

    iud_copper_t BOOLEAN DEFAULT FALSE,
    iud_hormonal_3_years BOOLEAN DEFAULT FALSE,
    iud_hormonal_5_years BOOLEAN DEFAULT FALSE,

    tubal_ligation BOOLEAN DEFAULT FALSE,
    vasectomy BOOLEAN DEFAULT FALSE,

    fam_standard_days BOOLEAN DEFAULT FALSE,
    fam_lam BOOLEAN DEFAULT FALSE,
    fam_two_day BOOLEAN DEFAULT FALSE,

    postpartum_fp_timing VARCHAR(5),
    post_abortion_fp_timing VARCHAR(5),

    implant_removal_reason VARCHAR(5),
    implant_removal_timing VARCHAR(5),
    iud_removal_reason VARCHAR(5),
    iud_removal_timing VARCHAR(5),

    side_effects VARCHAR(200),

    cervical_screening_method VARCHAR(5),
    cervical_cancer_status VARCHAR(5),
    cervical_cancer_treatment VARCHAR(5),
    breast_cancer_screening VARCHAR(10),

    screened_for_sti BOOLEAN,

    referral_number VARCHAR(50),
    referral_reason VARCHAR(300),

    remarks TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_fp_reg_facility_date ON fp_registrations(facility_id, visit_date);
CREATE INDEX IF NOT EXISTS idx_fp_reg_client_number ON fp_registrations(client_number) WHERE client_number IS NOT NULL;

-- Audit logs (append-only, no soft delete)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    facility_id UUID,
    action VARCHAR(30) NOT NULL,
    entity VARCHAR(50),
    entity_id VARCHAR(50),
    old_values TEXT,
    new_values TEXT,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    detail VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
