#!/bin/bash

#Create some bogus index by different name and different date with ElasticSearch Rest API.

function push_index()
{
    for line in `cat index_days.list`; do
       dayseq=${line%%=*}
       daystr=${line##*=}
       cur_uuid=$(cat /proc/sys/kernel/random/uuid)
       create_operation_index $daystr
       create_index $daystr project.myapp-dev
       create_index $daystr project.dev-${dayseq}
       create_index $daystr project.yourdev-${dayseq}
       create_index $daystr project.proddev-${dayseq}
       create_index $daystr prod-${dayseq}
    done
}

index_snapshot="index_snap.`date +%m%d%H%M%S`"
function print_index()
{
   for line in `cat index_days.list`; do
       daystr=${line##*=}
       eval $curl_via_cmd/_cat/indices?v  | awk /$daystr/{'print $3 '} |sort  |tee -a $index_snapshoft
       echo "#" |tee -a $index_snapshoft
   done
}
#############Main###########################

if [[ $# == 0 ]]; then
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
   push_index
   print_index
else
   print_index
fi
