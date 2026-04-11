-- services/payment-service/migrations/0001_init.down.sql
DROP INDEX IF EXISTS idx_payments_patient_id;
DROP INDEX IF EXISTS idx_payments_appointment_id;
DROP TABLE IF EXISTS payments;
