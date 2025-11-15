CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  email VARCHAR(255) UNIQUE NOT NULL,
  password TEXT NOT NULL,
  phone_number VARCHAR(50),
  is_verified BOOLEAN DEFAULT FALSE,
  kyc_status VARCHAR(50) DEFAULT 'pending',
  preferred_fiat VARCHAR(10) DEFAULT 'NGN',
  default_currency VARCHAR(10) DEFAULT 'NGN',
  last_login TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);