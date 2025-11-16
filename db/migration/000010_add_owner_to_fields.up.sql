ALTER TABLE "fields" ADD COLUMN "responsible_id" INT;
COMMENT ON COLUMN "fields"."responsible_id" IS 'Ответственный';

ALTER TABLE "fields" 
ADD CONSTRAINT fk_department_responsible
FOREIGN KEY (responsible_id) REFERENCES users(id);