Ancient Places Country Search - Project README
==============================================

About
-----

This project leverages archaeological data from [Pleiades][1] and
spatial data on country boundaries from [Natural Earth][2] to make it
possible to search for places of historical significance by country
name. Pleiades offers robust search already, but this is one missing
feature.

The two data sets are combined by performing [spatial joins][3] using
the [PostGIS][4] extension to the [PostgreSQL][5]
database. User-facing search functionality is provided via a web
application written in Go that uses the [Echo][6] web framework.

This project was undertaken mainly for personal learning, and is
publicly shared so that anyone interested in the underlying data or
the technologies used may benefit. It is a work in progress.

If you happen to be viewing this file offline, you may obtain the
latest source from the public Git repository:

<https://github.com/benghancock/ancient-places>

This project is currently deployed at the following URL:

<https://ancient-places-search-25vto.ondigitalocean.app/>

For more about this project, see the file [`doc/about.txt`](doc/about.txt).


Building & Dependencies
-----------------------

To build the database component of this project, you will need to
install or have access to a PostgreSQL database server; you will also
need to install the PostGIS extension, and some other related
command-line tools. See the file [`doc/guide.txt`](doc/guide.txt) for
a full tutorial on retrieving the source data and setting up the
database.

In order to build and run the web application, you will need a recent
version of the Go programming language (1.19 or later recommended). To
build from source, use the command `go build` inside this
repository. This will create a binary called `ancient-places`, which
will then run and serve the project over HTTP.

To build the HTML versions of the documentation, you will need
[pandoc][7], and optimally, GNU Make. Although the HTML docs may be
generated without Make, the `Makefile` in this repository includes
commands to do this. With Make installed, you should be able to build
the whole project (aside from the database) just by running:

```
$ make
```

Copying
-------

This project is free software. The source code and binaries may be
used in accordance with the `LICENSE` file in this repository. The
documentation for this project may be used under the terms of the
[CC-BY-SA 4.0][8] license.

This repository does not include copies of the underlying data from
either Pleiades or Natural Earth. For information on copying or
remixing data from those sources, see their respective websites.


Contact
-------

For questions or comments about this project, please write an email to
Ben Hancock at `mail [at] benghancock (dot) com`.


[1]: https://pleiades.stoa.org
[2]: https://www.naturalearthdata.com
[3]: https://www.postgis.net/workshops/postgis-intro/joins.html
[4]: https://postgis.net/
[5]: https://www.postgresql.org/
[6]: https://echo.labstack.com/
[7]: https://pandoc.org/
[8]: https://creativecommons.org/licenses/by-sa/4.0/
