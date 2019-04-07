#! /bin/bash

cd bin

nohup ./master &
nohup ./worker &

echo "deploy finished"
