================================
Ancient Places - Database Setup
================================

:Author:  Ben Hancock
:Date:    Summer 2023

Overview
--------

This document covers how to fetch the data to build the Ancient Places
database, and setting up the database itself. This project uses PostgreSQL
(version 14 or later is assumed) and the PostGIS extension. This tutorial
assumes that you are setting up the database yourself, rather than using a
managed service.


Retrieving the Data
-------------------

As noted in the README, this project utilizes data from two main sources:
`Pleiades`_ for data on archaelogical sites, and  geographical data from
`Natural Earth`_. This section covers how to retrieve and prepare the data
prior to setting up the database that will eventually store it for use with the
application.

Pleiades offers its data in a variety of formats, all of which are available at
this URL: https://pleiades.stoa.org/downloads

For this project, we will use the `GIS package`_, and specifically the
"places*" tables that it contains. For easy retrieval of the necessary tables,
use the ``fetch_pleiades_places.sh`` script in the ``scripts/`` directory,
found in the root of this project.

.. _Pleiades: https://pleiades.stoa.org/
.. _Natural Earth: https://www.naturalearthdata.com/
.. _GIS package: https://atlantides.org/downloads/pleiades/gis/

