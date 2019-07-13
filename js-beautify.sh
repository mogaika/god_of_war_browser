#!/usr/bin/env bash

dir="$(pwd)/web/data/static"
for file in $(find ${dir}|grep gow|grep .js); do
  js-beautify -r $file;
done

