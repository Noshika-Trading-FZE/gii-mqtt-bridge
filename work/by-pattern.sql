

CREATE OR REPLACE FUNCTION pix.i_get_object_by_property_pattern(v_group_name character varying, v_property character varying, v_value character varying) RETURNS uuid
    LANGUAGE sql STABLE STRICT PARALLEL SAFE
    AS $_$
    SELECT object_properties.object_id from pix.object_properties
        WHERE
                  group_name = $1 and
                  property = $2 and
                  value ~ $3 LIMIT 1
$_$;

CREATE OR REPLACE FUNCTION pix.create_control_execution_stealth_by_property_pattern(
            group_name character varying,
            property character varying,
            value character varying,
            name character varying,
            params json,
            enrich_with pix.property[]) RETURNS boolean
    LANGUAGE plpgsql STRICT PARALLEL SAFE
    AS $$
DECLARE
	v_record pix.control_executions;
	v_enrich_with ALIAS FOR enrich_with;
	v_params ALIAS FOR params;
	v_extra_props jsonb;
	v_default_params jsonb;
	
BEGIN
	v_record.object_id = pix.i_get_object_by_property_pattern(group_name, property, value);

	IF (v_record.object_id IS NULL) THEN
		RETURN false;
	END IF;
	
	v_record.helper_params = pix.i_get_rpc_helper_params(v_record.object_id, v_enrich_with);	
	v_record.name = name;
	v_record.params = params;
	v_record.type = 'RPC_STEALTH';
	v_record.controller = pix.i_resolve_controller(v_record);
	v_record.created_at = now();
	v_record.caller_id = (pix.i_current_user_id())::uuid;

	IF v_record.controller IS NULL THEN 
		RETURN false;
--		RAISE EXCEPTION 'Stealth RPC: Couldn''t resolve controller for RPC. Aborting.';
	END IF;

	CALL pix.do_publish('insert','control_executions',v_record);

--	res = pg_notify('pix:controls:','test'); 

    RETURN true;


END;
$$;


ALTER FUNCTION pix.create_control_execution_stealth_by_property_pattern(group_name character varying, property character varying, value character varying, name character varying, params json, enrich_with pix.property[]) OWNER TO postgres;
COMMENT ON FUNCTION pix.create_control_execution_stealth_by_property_pattern(group_name character varying, property character varying, value character varying, name character varying, params json, enrich_with pix.property[]) IS 'Initiate execution of Stealth RPC/Control where object is identified by GroupName, Property, Value. ';


