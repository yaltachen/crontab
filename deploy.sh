#! /bin/bash

cd bin

nohup ./master --config ../config/master.json &
nohup ./worker --config ../config/worker.json &

echo "deploy finished"
