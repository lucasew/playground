-- migrate:up
CREATE OR REPLACE FUNCTION handle_users_audit()
    RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

-- Trigger for INSERT operations.
CREATE TRIGGER users_insert_audit AFTER INSERT ON users FOR EACH ROW EXECUTE FUNCTION handle_users_audit();

-- Trigger for UPDATE operations.
CREATE TRIGGER users_update_audit AFTER UPDATE ON users FOR EACH ROW EXECUTE FUNCTION handle_users_audit();

-- Trigger for DELETE operations.
CREATE TRIGGER users_delete_audit AFTER DELETE ON users FOR EACH ROW EXECUTE FUNCTION handle_users_audit();
 
-- migrate:down
DROP TRIGGER users_insert_audit ON users;
DROP TRIGGER users_update_audit ON users;
DROP TRIGGER users_delete_audit ON users;
DROP FUNCTION handle_users_audit;

