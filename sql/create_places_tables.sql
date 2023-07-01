-- Create the general "places" table for the Pleiades data
CREATE TABLE IF NOT EXISTS places (
	created timestamp with time zone,
	description text,
	details text,
	provenance varchar(500),
	title varchar(500),
	uri varchar(1000),
	id bigint,
	representative_latitude decimal,
	representative_longitude decimal,
	bounding_box_wkt text
);

-- TODO! Create place_types
-- TODO! Create places_place_types
