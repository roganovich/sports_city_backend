ALTER TABLE "fields" ADD COLUMN "slug" varchar;
ALTER TABLE "fields" ALTER COLUMN slug SET NOT NULL;
COMMENT ON COLUMN "fields"."slug" IS 'Slug';