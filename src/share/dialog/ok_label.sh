#!/bin/bash
# useradd1.sh - A simple shell script to display the form dialog on screen
# set field names i.e. shell variables
shell=""
groups=""
user=""
home=""

# open fd
exec 3>&1

# Store data to $VALUES variable
VALUES=$(dialog --ok-label "Submit" \
	  --backtitle "Linux User Managment" \
	  --title "Useradd" \
	  --form "Create a new user" \
15 50 0 \
	"Username:" 1 1	"$user" 	1 10 10 0 \
	"Shell:"    2 1	"$shell"  	2 10 15 0 \
	"Group:"    3 1	"$groups"  	3 10 8 0 \
	"HOME:"     4 1	"$home" 	4 10 40 0 \
2>&1 1>&3)

# close fd
exec 3>&-

# display values just entered
echo "$VALUES"
