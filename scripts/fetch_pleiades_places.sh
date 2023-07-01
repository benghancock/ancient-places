#! /bin/bash

# Retrieve the Pleiades GIS package and extract the ``places.csv`` file

PLEIADES_GIS_URL='https://atlantides.org/downloads/pleiades/gis/pleiades_gis_data.zip'
TEMPFILE=$(mktemp pleiades_gis_data.zip.XXXXXXX)
TEMPDIR=$(mktemp -d pleiades_data.XXXXXX)
ERRSTAT=0

echo "Fetching data from Pleiades ..."
curl --silent --show-error  -o $TEMPFILE $PLEIADES_GIS_URL

if [ $? != 0 ]; then
	echo "Error retrieving data!"
	ERRSTAT=1
else
	echo "Extracting files ..."
	unzip -q -d $TEMPDIR $TEMPFILE
	mv $TEMPDIR/place*.csv .
fi

echo "Cleaning up ..."
rm -rf $TEMPDIR
rm $TEMPFILE

echo "All done!"
exit $ERRSTAT
