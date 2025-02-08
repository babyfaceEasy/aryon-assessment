-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tenants table if it does not exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'connectors') THEN
        CREATE TABLE connectors (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            workspace_id varchar(255) NOT NULL,
            default_channel_id varchar(255) NOT NULL,
            created_at TIMESTAMPTZ DEFAULT NOW(),
            updated_at TIMESTAMPTZ DEFAULT NOW()
        );
    END IF;
END $$;
