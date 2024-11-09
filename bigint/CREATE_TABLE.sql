CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE "user"(
    u_id BIGINT PRIMARY KEY,
    avatar_file_id BIGINT
);
CREATE TABLE "file"(
    file_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    file_uuid UUID DEFAULT uuid_generate_v4(),
    file_extension TEXT,
    created_at TIMESTAMPTZ
);
ALTER TABLE "user"
ADD CONSTRAINT user_file_fk FOREIGN KEY (avatar_file_id) REFERENCES "file"(file_id) ON UPDATE CASCADE ON DELETE CASCADE;
CREATE INDEX idx_user_avatar_file_id ON "user"(avatar_file_id);
