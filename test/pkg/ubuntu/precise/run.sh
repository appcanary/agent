#!/bin/bash
provision=$(docker run appcanary-debian-jessie /root/test.sh)
echo "$provision"
okcount=`echo "$provision"|grep OKAC|wc -l`
if [ "$okcount" -eq "4" ]
then
  echo TEST PASS
else
  echo FAIL
fi
