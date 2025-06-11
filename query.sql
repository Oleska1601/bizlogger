CREATE TABLE if not exists business_logs  (
id SERIAL PRIMARY KEY,
timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
event_type VARCHAR(10) NOT NULL, -- CREATE, UPDATE, DELETE
entity VARCHAR(20) NOT NULL, -- user, context
username VARCHAR(36) NOT NULL,
user_role VARCHAR(2) NOT NULL, -- SA, CA
context VARCHAR(50),
entity_id VARCHAR(50) NOT NULL,
old_value JSONB,
new_value JSONB,
description TEXT
);