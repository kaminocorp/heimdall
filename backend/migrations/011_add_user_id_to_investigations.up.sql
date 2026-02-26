ALTER TABLE investigations
    ADD COLUMN user_id UUID NOT NULL
    DEFAULT '00000000-0000-0000-0000-000000000000'
    REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE investigations ALTER COLUMN user_id DROP DEFAULT;

CREATE INDEX idx_investigations_user_id ON investigations(user_id);
