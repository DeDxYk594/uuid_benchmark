CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE "user"(
    u_id BIGINT PRIMARY KEY,
    avatar_file_uuid UUID
);
CREATE TABLE "file"(
    file_uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_extension TEXT,
    created_at TIMESTAMPTZ
);
ALTER TABLE "user"
ADD CONSTRAINT user_file_fk FOREIGN KEY (avatar_file_uuid) REFERENCES "file"(file_uuid) ON UPDATE CASCADE ON DELETE CASCADE;
CREATE INDEX idx_user_avatar_file_uuid ON "user"(avatar_file_uuid);
