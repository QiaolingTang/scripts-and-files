#!/bin/bash

#Create some bogus index by different name and different date with ElasticSearch Rest API.

cat <<EOF >index_days.list
d1=$(date -d "-1 day" +%Y.%m.%d)
d6=$(date -d "-6 day" +%Y.%m.%d)
d7=$(date -d "-7 day" +%Y.%m.%d)
d8=$(date -d "-8 day" +%Y.%m.%d)
d13=$(date -d "-13 day" +%Y.%m.%d)
d14=$(date -d "-14 day" +%Y.%m.%d)
d15=$(date -d "-15 day" +%Y.%m.%d)
d29=$(date -d "-29 day" +%Y.%m.%d)
d30=$(date -d "-30 day" +%Y.%m.%d)
d31=$(date -d "-31 day" +%Y.%m.%d)
d60=$(date -d "-60 day" +%Y.%m.%d)
EOF

echo "created days list"
