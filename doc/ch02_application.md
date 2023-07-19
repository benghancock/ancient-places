---
title: Ancient Places - Building the Search Application
author: Ben Hancock
date: Summer 2023
---

# Overview

In this document, I will cover the process of building and running the
application that talks to the `archaia` database and allows a user to
search for Pleiades places by country name. If you're following along,
you'll want to make sure you have a recent version of Go installed on
the system where you're working.


# Configuring a Database User

In the first section of our documentation about setting up the
database, we reviewed the security model for data access and discussed
the creation of roles. The idea is fairly simple: have one role for
administering the database, and another for the application, which
will essentially have read only privileges to the tables or views it
needs. Let's create the application user now:

```
$ sudo -u postgres psql
# ... prompted for password ...
psql (14.3)
Type "help" for help.

postgres=# CREATE ROLE archaia_app_user WITH LOGIN ENCRYPTED PASSWORD 's00per-s3cret';

```

As a refresher: In Chapter 1, we created a "materialized view" of the
results of a query with all the data we should need for our
application. We called this object `countries_places`; giving our app
user `SELECT` privileges on this view should be sufficient:

```
postgres=# \c archaia
You are now connected to database "archaia" as user "postgres".
archaia=# GRANT SELECT ON countries_places TO archaia_app_user;
GRANT
```

Let's test this out by disconnecting and reconnecting as the app user:

```
$ psql -d archaia -U archaia_app_user
psql: error: connection to server on socket "/var/run/postgresql/.s.PGSQL.5432" failed: FATAL:  Peer authentication failed for user "archaia_app_user"
```

Ah! We haven't yet configured our database server to accept
password-based authentication. We can do this in Postgres using the
file [`pg_hba.conf`]. Where this is on your system will vary by
install; on Fedora it should be at `/var/lib/pgsql/data/pg_hba.conf`
(see [more Fedora docs here]). The file itself is pretty well
commented, but consult the Postgres documentation for the specifics on
the formatting. We'll add a line like this:

```
# "local" is for Unix domain socket connections only
local   archaia         archaia_app_user                        md5
```

NB: Order matters in this file, so we'll want to put this line *above*
other `local` rules. Once we've edited the configuration, make sure to
restart the database server, and then try logging in again -- ensuring
that our permissions are correctly limited.

```
$ sudo systemctl restart postgresql
$ psql -d archaia -U archaia_app_user
Password for user archaia_app_user:
psql (14.3)
Type "help" for help.

archaia=> SELECT count(*) FROM places;
ERROR:  permission denied for table places
archaia=> SELECT count(*) FROM countries_places;
 count
-------
 35774
(1 row)
```

Looking good! Now we're ready to move on to building the application,
and connecting to the DB from our code.

[`pg_hba.conf`]: https://www.postgresql.org/docs/14/auth-pg-hba-conf.html
[more Fedora docs here]: https://docs.fedoraproject.org/en-US/quick-docs/postgresql/#pg_hba.conf
