#!/bin/sh

git pull
git add --all
git commit -am "${1}"
git push
