---
title: Ancient Places - Database Setup
author: Ben Hancock
date: Summer 2023
---

# Overview

This document covers how to fetch the data to build the Ancient Places
database, and setting up the database itself. This project uses
PostgreSQL (version 14 or later is assumed) and the PostGIS
extension. This tutorial assumes that you are setting up the database
yourself, rather than using a managed service.


# Retrieving the Data

As noted in the README, this project utilizes data from two main
sources: [Pleiades] for data on archaelogical sites, and geographical
data from [Natural Earth]. This section covers how to retrieve and
prepare the data prior to setting up the database that will eventually
store it for use with the application.

Pleiades is a public repository or “gazetteer” of geographic
information about the ancient world. It offers its data in a variety
of formats, all of which are available at this URL:
<https://pleiades.stoa.org/downloads>

For this project, we will use the [GIS package], and specifically the
`"places*"` tables that it contains. For easy retrieval of the
necessary tables, use the `fetch_pleiades_places.sh` script in the
`scripts/` directory, found in the root of this project.

Here is a useful excerpt about "places" from the Pleiades website:

> Pleiades places are the primary organizational construct of the
> gazetteer. They are conceptual entities: the term "place" applies to
> any locus of human attention, material or intellectual, in a
> real-world geographic context.

Note: The document mentions that places are "entirely abstract, conceptual
entities [...] objects of thought, speech, or writing, not tangible,
mappable points on the earth’s surface." This may introduce some
wrinkles into our project further down the line, but we'll visit those
when we get to them.

Running this script will leave three files in the working directory:

* `places.csv`: Pleiades Places.
* `place_types.csv`: terms from the Place Types
* `places_place_types.csv`: matches place ids (join to places.csv:id)
  to placetype ids (join to place_types.csv:key).

Once we've retrieved these, we're ready for the next step.

[Pleiades]: https://pleiades.stoa.org/
[Natural Earth]: https://www.naturalearthdata.com/
[GIS package]: https://atlantides.org/downloads/pleiades/gis/

# Setting up the Database

## Configuring Roles

When starting out configuring your database, it's a good idea to think
about your security model -- even if this is just a hobby project with
public data. Again, in this example I will assume that you are setting
up the database yourself, and that we will be running both the
Postgres server and the application on the same host. If you are using
a managed service in the cloud, this will almost certainly not be the
case, so your mileage may vary.

Our security model will be pretty basic: we will create one role to
administer the database (create the tables, etc.), and another role
for the application with `SELECT`-only privileges to read the tables
or views that it needs.

For this document, I will leave aside how to install and set up the
database server itself. This likely varies depending on your operating
system anyway; I built this project on a local installation of Fedora
Linux on my laptop, so if you’re in a similar environment, you may
find this tutorial helpful:
<https://docs.fedoraproject.org/en-US/quick-docs/postgresql/>

Once you have your database server installed and running, connect using the
`psql` command line tool. You will first need to to this as the Postgres
"superuser", named `postgres`. Below is an example session; note that the
statements that were executed are echoed after they succeed:

```
$ sudo -u postgres psql
# ... prompted for password ...
psql (14.3)
Type "help" for help.

postgres=# CREATE DATABASE archaia;
CREATE DATABASE
postgres=# CREATE ROLE archaia_admin NOINHERIT;
CREATE ROLE
postgres=# ALTER DATABASE archaia OWNER TO archaia_admin;
ALTER DATABASE
postgres=# GRANT archaia_admin TO {your user here};
GRANT ROLE
```

Here we are making use of Postgres' [role-based permissions model], by
creating a database and a dedicated admin role for that database that
has no other special permissions. In the final statement, we make our
user (that is, the username you use on your operating system) a
*member* of the `archaia_admin` group, so that we can administer this
database from our regular user account without creating a new special
password. We could also make others members of this group, and remove
them as appropriate.

Now, we are ready to exit our superuser session and log back in:

```
postgres=# \q
$ psql -d archaia
psql (14.3)
Type "help" for help.

archaia=>
```

[role-based permissions model]: https://www.postgresql.org/docs/14/user-manag.html

## Creating Tables and Loading in Data

Once we have done this, we can now create the necessary tables to hold
the "places" data we fetched from Pleiades. Please refer to the file
`create_places_tables.sql` in the ``sql/`` directory of this project.

With the tables created, we can now load data into them using the
`COPY` statement, or `psql`'s analogous `\copy` command (the latter of
which typically bypasses permissions issues).

```
archaia=> \copy places from '/path/to/places.csv' with (format csv, header);
COPY 38953
archaia=> \copy places_types from '/path/to/place_types.csv' (format csv, header);
COPY 180
archaia=> \copy places_place_types from '/path/to/places_place_types.csv' (format csv, header);
ERROR:  insert or update on table "places_place_types" violates foreign key constraint "places_place_types_place_type_fkey"
DETAIL:  Key (place_type)=(quarry-group) is not present in table "places_types".
```

In our attempt to load data into the last table, we've run into an error; it
looks like although we have a place with a "quarry-group" type, that type
isn't in our "places_types" table. We can dig into this issue with a few
shell commands on the CSV data.

First, let's see how many records this impacts:

```
$ cat places_place_types.csv | grep 'quarry-group' | wc -l
3
```

Ok, just three records -- not bad. We can drop those, or maybe we can find
a close-enough type that would be appropriate?

```
$ cat place_types.csv | grep '^quarry'
quarry,quarry,A quarry as defined by the Getty Art and Architecture Thesaurus: Open-air excavations from which stone for building or other purposes is or has been obtained by cutting or blasting.
```

That seems close enough. Let's update those three records in our data file
and try loading again:

```
$ sed 's/quarry-group/quarry/' places_place_types.csv > places_place_types1.csv
$ fg
psql -d archaia

archaia=> \copy places_place_types from '/path/to/places_place_types1.csv' (format csv, header);
ERROR:  insert or update on table "places_place_types" violates foreign key constraint "places_place_types_place_type_fkey"
DETAIL:  Key (place_type)=(labeled feature) is not present in table "places_types".
```

Darn, we've hit another missing key. This one affects a higher number
of records (more than 250), and there's nothing obviously analogous in
the `place_types.csv` file. So let's drop the constraint, and move on:

```
archaia=> ALTER TABLE places_place_types
archaia-> DROP CONSTRAINT places_place_types_place_type_fkey;
ALTER TABLE
archaia=> \copy places_place_types from '/home/bgh/projects/code/ancient-places/places_place_types1.csv' (format csv, header);
COPY 43963
```

Now that we're set up, let's test our table setup with a query. (At
this point, you may want to consider switching to a tool like
[`pgadmin`] for running queries and viewing the output.)

``` sql
SELECT
  places.title,
  places.id,
  places_place_types.place_type,
  places_types.definition
FROM places
  LEFT JOIN places_place_types
  ON places.id = places_place_types.place_id
  LEFT JOIN places_types
  ON places_place_types.place_type = places_types.key
ORDER BY places.title ASC
LIMIT 200;
```

# Creating PostGIS 'Geometry' Columns

Next, we'll want to make sure that we have the PostGIS extension for
Postgres installed and enabled; this will allow us to work with the
geographical data in the Pleiades dataset. You can enable the
extension using the `CREATE EXTENSION` statement.

The first time I tried this, though, I ran into an error:

```
ancient_places=> CREATE EXTENSION postgis;
ERROR:  could not open extension control file "/usr/share/pgsql/extension/postgis.control": No such file or directory
```

I found that several standalone packages were provided for PostGIS
by Fedora's package manager, and installed them:

```
$ sudo dnf install postgis postgis-docs postgis-upgrade postgis-utils
```

Now, running the create extension statement works, though note that
you must be the database superuser (by default, `postgres`) in order
to successfully execute it. To check the version of PostGIS installed,
execute this statement:

``` sql
SELECT postgis_full_version();
```

Throughout this document, I'm running PostGIS 3.2.2.

The next thing for us to do is to create columns for the data using one
of PostGIS' supported geospatial data types. There are two main types
to choose from: `geography` and `geometry`. The PostGIS documentation
on [why to choose one over the other] is helpful on this topic, but to briefly
paraphrase: the `geography` type is appropriate for geographically dispersed
data, whereas `geometry` is generally appropriate for more geographically
compact data. That said, PostGIS offers some useful functions for dealing
with `geometry` data, and casting from one to the other is trivial.

::: Note

**A Word (or Two) About Spatial Reference Systems**

One important concept when working with spatial data is the idea of
*coordinates systems*, or *spatial reference systems*. Explained
simply, these are systems that humans can use to reflect the
location of places on the earth on a map. As expressed more artfully
by the [PostGIS documentation](https://www.postgis.net/workshops/postgis-intro/projection.html):

> The earth is not flat, and there is no simple way of putting it down
> on a flat paper map (or computer screen), so people have come up
> with all sorts of ingenious solutions, each with pros and cons. Some
> projections preserve area, so all objects have a relative size to
> each other; other projections preserve angles (conformal) like the
> Mercator projection; some projections try to find a good
> intermediate mix with only little distortion on several
> parameters. Common to all projections is that they transform the
> (spherical) world onto a flat Cartesian coordinate system, and which
> projection to choose depends on how you will be using the data.

The thing to underline here is that, when working with geographic
data, you should know which spatial reference system it uses. In
PostGIS and other GIS systems, these are referred by their spatial
reference identifier (SRID). More on that here:

<https://postgis.net/workshops/postgis-intro/loading_data.html#srid-26918-what-s-with-that>

Setting the appropriate SRID ensures that when using spatial functions
to calculate distance, etc., your results will be correct. This is
even more important when working with data that utilizes different
spatial reference systems. The most common SRID for geopgraphic
coordinates is SRID 4326, which corresponds to “longitude/latitude on
the WGS84 spheroid” [^1]. Luckily that's what all of our data uses
throughout these exercises.

:::

We will use the representative lat/lon coordinates in our `places` table
in order to construct a column of the `geography` data type.

``` sql
ALTER TABLE places
ADD COLUMN repr_geog geography(POINT, 4326);

UPDATE places
SET repr_geog =
  ST_SetSRID(
    ST_MakePoint(representative_longitude, representative_latitude),
    4326)::geography;
```

[why to choose one over the other]: https://www.postgis.net/workshops/postgis-intro/geography.html#why-not-use-geography

[^1]: https://www.postgis.net/workshops/postgis-intro/projection.html#transforming-data

# Importing Data from Natural Earth

Now that we have our data on archaeological sites in the shape we want
it, we need to import data on modern-day geopolitical borders in order
to support our goal of enabling search by country. For this, I found
the "Cultural Vectors" data from Natural Earth to fit the bill:

<https://www.naturalearthdata.com/downloads/10m-cultural-vectors/>

Specifically, I downloaded the "Admin 0 - Countries" dataset. Country
boundaries are, of course, a political construct and may be fluid or
disputed. To this point, Natural Earth notes that it "shows de facto
boundaries by default according to who controls the territory, versus
de jure." (It also offers "point-of-view" border data.)

This dataset is offered as a [shapefile], and so to import it into
Postgres, we'll want to use a tool called `shp2pgsql`. This is a
command-line tool that is probably included with the `postgis-client`
package in your package manager. This tool has a lot of command-line
options, and it's worth exploring them by running `shp2pgsql -?`. But
for our use-case, no flags are necessary:

```
$ wget 'https://www.naturalearthdata.com/http//www.naturalearthdata.com/download/10m/cultural/ne_10m_admin_0_countries.zip'
$ unzip ne_10m_admin_0_countries.zip
$ shp2pgsql ne_10m_admin_0_countries.shp countries_political | psql -d archaia
```

You'll see a lot of output fly by in your terminal as the data is
loaded into a new table in the `archaia` database, called
`countries_political`. Back in your query tool (either pgadmin or
psql), you can get information on this new table like so:

``` sql
SELECT column_name
FROM information_schema.columns
WHERE table_name = 'countries_political';
```

There are *a lot* of columns in this dataset, but there are really
only two that we need to be concerned with for now: `sovereignt`
and `geom`, which are the names of the countries (in English) and
their geometries, respectively. Take a quick look:

``` sql
SELECT sovereignt, geom
FROM countries_political
ORDER BY sovereignt ASC
LIMIT 10;
```

We're now almost ready to do a join with our Pleiades `places` table, in
order to place the sites in the modern-day borders where they can now
be found. But a couple additional considerations need to be made first.

First, we should consider that these places may often *not* fall
neatly within current-day geopolitical borders. It's always possible
that they could fall _on_ a border, and we would like to at least try
to accomodate this scenario.

This means that our places may have a one-to-many relationship with
our countries data, and we should choose a join method that would
support this. Furthermore, in line with the Pleiades note that places
are often "conceptual" rather than actual mappable points on the
earth, it follows that many "places" will actually not have
coordinates at all, and thus won't fall into a "country" as such --
though they may have cultural ties to actual, geographic locations.

In practical terms, this means that we probably want to use the
PostGIS function [`ST_Intersects`] to perform our join, rather than
[`ST_Contains`], since we will want to join places on all of the
countries that they may touch. Even this is not perfect ; Pleiades'
cautions that often the coordinates given may just be the centroid of
a very large bounding box covering a large area. Some of the places
are also along roads or routes that traverse long distances.

Before we give this a try, we'll also want to create indexes on our
spatial data columns, in order to make our query reasonably
performant, and also make sure that we set the SRID for the `geom`
column from the Natural Earth Data.

``` sql
CREATE INDEX countries_geom_idx ON countries_political
USING GIST (geom);

CREATE INDEX places_repr_geog_idx ON places
USING GIST (repr_geog);
```


[shapefile]: https://en.wikipedia.org/wiki/Shapefile
[`ST_Intersects`]: https://postgis.net/docs/manual-dev/en/ST_Intersects.html
[`ST_Contains`]: https://postgis.net/docs/manual-dev/en/ST_Contains.html

# Creating Views

TODO
