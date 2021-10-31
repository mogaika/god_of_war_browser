#!/usr/bin/env bash

dir="$(pwd)/web/data/static/js/"
for file in $(find ${dir}|grep .js|grep -v vendor); do
  js-beautify -r $file;
done

