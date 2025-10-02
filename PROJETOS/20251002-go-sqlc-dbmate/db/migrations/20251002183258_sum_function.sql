-- migrate:up
CREATE FUNCTION sum(integer, integer)
    RETURNS integer
    AS 'select $1 + $2;'
    LANGUAGE SQL;


-- migrate:down
DROP FUNCTION sum;
