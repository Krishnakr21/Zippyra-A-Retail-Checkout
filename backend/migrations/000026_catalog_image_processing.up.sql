-- Add image_processing_status to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_processing_status VARCHAR(20) DEFAULT 'PROCESSED';
