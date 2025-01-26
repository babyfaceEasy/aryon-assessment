-- Create tenant table (just for the example)
CREATE TABLE tenants
(
    tenant_id   UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_name VARCHAR(255) UNIQUE NOT NULL
);



