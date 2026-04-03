-- services/patient-service/migrations
DROP INDEX IF EXISTS idx_patients_email;
DROP INDEX IF EXISTS idx_patients_user_id;
DROP TABLE IF EXISTS patients;
