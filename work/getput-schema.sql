
CREATE OR REPLACE FUNCTION pix.export_b64_schema(schema_id uuid) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
        v_schema text;
BEGIN
    v_schema = pix.export_schema(schema_id)::text;
    RETURN encode(v_schema::bytea, 'base64');
END;
$$;


CREATE OR REPLACE FUNCTION pix.import_b64_schema(b64_schema text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_schema text;
    j_schema json;
BEGIN
    -- RAISE WARNING '%', b64_schema;
    v_schema = convert_from(decode(b64_schema, 'base64'), 'UTF8');
    -- RAISE WARNING '%', v_schema;
    -- j_schema = v_schema::json;
    -- RAISE WARNING '%', j_schema;
    RETURN pix.import_schema(v_schema::json);
END;
$$;
