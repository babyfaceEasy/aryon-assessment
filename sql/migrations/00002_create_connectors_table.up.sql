-- Create connectors table if it does not exist
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

        -- add index to connectors table to improve performance
        CREATE INDEX idx_connectors_id ON connectors(id);
    END IF;
END $$;