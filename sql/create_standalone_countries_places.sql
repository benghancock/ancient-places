-- This query creates a standalone version of the 'countries_places' view,
-- which is useful when exporting the data and loading it to a remote db
-- for deployment.

CREATE TABLE IF NOT EXISTS countries_places (
    country_name       varchar(500),
	place_id           bigint,
	place_name         text,
	pleiades_uri       varchar(1000),
	place_type         varchar(50),
	place_type_def     text,
	descrip            text
);
