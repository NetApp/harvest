#!/bin/bash

exec 3>&1

title=$1
menu=$2
count=$3

declare -a ITEMS
for ((i = 0 ; i < $count ; i++)); do
    ITEMS+=($i+1)
    ITEMS+=($@[$i+3])
done

echo "items: $ITEMS"

VALUES=$(dialog --title "$title" \
    --menu "$menu" \
    10 30 $count \
    $ITEMS \
2>&1 1>&3)

errcode=$?

exec 3>&-

echo "values: $VALUES"
echo "exit code: $errcode"
