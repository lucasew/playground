-- Create "handle_users_audit" function
CREATE FUNCTION "public"."handle_users_audit" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
        IF (TG_OP = 'INSERT') THEN
            INSERT INTO users_audit_logs(operation_type, operation_time, new_value)
            VALUES (TG_OP, CURRENT_TIMESTAMP, row_to_json(NEW));
            RETURN NEW;
        ELSIF (TG_OP = 'UPDATE') THEN
            INSERT INTO users_audit_logs(operation_type, operation_time, new_value)
            VALUES (TG_OP, CURRENT_TIMESTAMP, row_to_json(OLD), row_to_json(NEW));
            RETURN NEW;
        ELSIF (TG_OP == 'DELETE') THEN
            INSERT INTO users_audit_logs(operation_type, operation_time, old_value)
            VALUES (TG_OP, CURRENT_TIMESTAMP, row_to_json(OLD));
            RETURN OLD;
        END IF;
        RETURN NULL;
    END;
$$;
-- Create trigger "users_delete_audit"
CREATE TRIGGER "users_delete_audit" AFTER DELETE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."handle_users_audit"();
-- Create trigger "users_insert_audit"
CREATE TRIGGER "users_insert_audit" AFTER INSERT ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."handle_users_audit"();
-- Create trigger "users_update_audit"
CREATE TRIGGER "users_update_audit" AFTER UPDATE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."handle_users_audit"();
-- Create "sum" function
CREATE FUNCTION "public"."sum" (integer, integer) RETURNS integer LANGUAGE sql AS $$ select $1 + $2; $$;
