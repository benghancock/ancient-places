-- Create a materialized view joining archaelogical data from Pleiades
-- and country border data from Natural Earth. Assumes that necessary
-- columns and indexes have already been created.
CREATE MATERIALIZED VIEW countries_places AS
SELECT
    cp.sovereignt as country_name,
    pt.place_id,
    pt.place_name,
    pt.pleiades_uri,
    pt.place_type,
    pt.place_type_def,
    pt.descrip
FROM (
    SELECT
        places.id as place_id,
        places.bb_geom as bb_geom,
        places.title as place_name,
        places.description as descrip,
        places.uri as pleiades_uri,
        places_place_types.place_type,
        places_types.definition as place_type_def
    FROM places
	    LEFT JOIN places_place_types
        ON places.id = places_place_types.place_id
        LEFT JOIN places_types
        ON places_place_types.place_type = places_types.key
    WHERE places.bounding_box_wkt IS NOT NULL
    ORDER BY places.title ASC
) as pt
LEFT JOIN
    countries_political cp
ON ST_Intersects(
    cp.geom,
	pt.bb_geom
)
ORDER BY cp.sovereignt ASC;
