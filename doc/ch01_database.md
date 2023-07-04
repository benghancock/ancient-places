---
title: Ancient Places - Database Setup
author: Ben Hancock
date: Summer 2023
---

# Overview

This document covers how to fetch the data to build the Ancient Places
database, and setting up the database itself. This project uses PostgreSQL
(version 14 or later is assumed) and the PostGIS extension. This tutorial
assumes that you are setting up the database yourself, rather than using a
managed service.


# Retrieving the Data

As noted in the README, this project utilizes data from two main sources:
[Pleiades] for data on archaelogical sites, and  geographical data from
[Natural Earth]. This section covers how to retrieve and prepare the data
prior to setting up the database that will eventually store it for use with the
application.

Pleiades is a public repository or “gazetteer” of geographic information about
the ancient world. It offers its data in a variety of formats, all of which are
available at this URL: <https://pleiades.stoa.org/downloads>

For this project, we will use the [GIS package], and specifically the
`"places*"` tables that it contains. For easy retrieval of the necessary tables,
use the `fetch_pleiades_places.sh` script in the `scripts/` directory,
found in the root of this project.

Running this script will leave three files in the working directory (described
below as in the Pleiades README file):

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

When starting out configuring your database, it's a good idea to think about
your security model -- even if this is just a hobby project with public data.
You don't want bad actors abusing your database. Again, in this example I will
assume that you are setting up the database yourself, and that we will be
running both the Postgres server and the application on the same host. If you
are using a managed service in the cloud, this will almost certainly not be the
case, so your mileage may vary.

Our security model will be pretty basic: we will create one role to administer
the database (create the tables, etc.), and another role for the application
with `SELECT`-only privileges to read the tables or views that it needs. For
this document, I will leave aside how to install and set up the database server
itself. This likely varies depending on your operating system anyway; I built
this project on a local installation of Fedora Linux on my laptop, so if you’re
in a similar environment, you may find this tutorial helpful:
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
creating a database and a dedicated admin role for that database that has no
other special permissions. In the final statement, we make our user (that is,
the username you use on your operating system) a *member* of the
`archaia_admin` group, so that we can administer this database from our
regular user account without creating a new special password. We could also
make others members of this group, and remove them as appropriate.

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

Once we have done this, we can now create the necessary tables to hold the
"places" data we fetched from Pleiades. Please refer to the file
`create_places_tables.sql` in the ``sql/`` directory of this project.

With the tables created, we can now load data into them using the `COPY`
statement, or `psql`'s analogous `\copy` command (the latter of which
typically bypasses permissions issues).

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
quarry,quarry,A quarry as defined by
the Getty Art and Architecture Thesaurus: Open-air excavations from which stone
for building or other purposes is or has been obtained by cutting or blasting.
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

Darn, we've hit another missing key. This one affects a higher number of
records (more than 250), and there's nothing obviously analogous in the
`place_types.csv` file. So let's drop the constraint, and move on:

```
archaia=> ALTER TABLE places_place_types
archaia-> DROP CONSTRAINT places_place_types_place_type_fkey;
ALTER TABLE
archaia=> \copy places_place_types from '/home/bgh/projects/code/ancient-places/places_place_types1.csv' (format csv, header);
COPY 43963
```

Now that we're set up, let's test our table setup with a query. (At this point,
you may want to consider switching to a tool like [`pgadmin`] for running
queries and viewing the output.)

```sql
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

# Importing Data from Natural Earth

# Creating Views
