-- Data definition statements for various "places" datasets from Pleiades

-- Create the general "places" table for the Pleiades data
CREATE TABLE IF NOT EXISTS places (
    created     timestamp with time zone,
    description text,
    details     text,
    provenance  varchar(500),
    title       varchar(500),
    uri         varchar(1000),
    id          bigint,
    representative_latitude     decimal,
    representative_longitude    decimal,
    bounding_box_wkt    text,
    CONSTRAINT place_id PRIMARY KEY (id)
);

-- Create a table for the various place types
CREATE TABLE IF NOT EXISTS places_types (
    key         varchar(50),
    term        text,
    definition  text,
    CONSTRAINT place_type_key PRIMARY KEY (key)
);

-- Create a fact table to join places to their place type definitions
CREATE TABLE IF NOT EXISTS places_place_types (
    place_id    bigint      REFERENCES places (id),
    place_type  varchar(50) REFERENCES places_types (key)
);
