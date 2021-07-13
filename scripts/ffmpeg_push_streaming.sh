#!/bin/bash

while ((1))
do
  curl http://localhost:8090/control/get?room=movie
  ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk
  sleep 3
done
