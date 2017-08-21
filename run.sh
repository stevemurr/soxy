#!/bin/bash

# This is just for convenience

resample()
{
    SRC=$1
    DST=$2
    TIME_RESULTS="$DST/times.txt"
    if [ -d $DST ]; then
        echo "Removing $DST ..."
        rm -rf $DST
    fi
    mkdir -p $DST
    for file in $SRC/*.wav;
        do 
            BASE=$(basename $file)
            TIME_RESULTS="$DST/$BASE.txt"
            echo "Converting $file ..."
            echo "Filename: $file" > $TIME_RESULTS
            cp $file $DST/original_$(basename $file)
            { time ./soxy -in $file -out $(echo $DST)/rez_$(echo $BASE) -c $3 ;} 2>> $TIME_RESULTS
    done;
}

# 1 in
# 2 out
# 3 config
resample $1 $2 $3