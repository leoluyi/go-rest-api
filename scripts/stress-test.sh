#!/usr/bin/env bash

ab -n 5000 -c 10 -k localhost:8080/v1/albums
