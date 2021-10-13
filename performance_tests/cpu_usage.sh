#!/bin/bash

while true; 
do 
    docker stats --no-stream | grep -v 'CONTAINER' | awk -v date="$(date +%T)" '{print $4, date}' >> cpu.txt; 
done;
