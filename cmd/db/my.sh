#!/bin/bash

term=$1

function usage() {
    echo "my <term0 term1|file.cnf>"

}

if [ -z $term ]; then
    usage
elif [ -f $term ]; then
  echo "file found $1"
        mycli --defaults-file=$1
elif [ -n $term ]; then
  echo "my $1"
  cnf=$($HOME/bin/db findone $@)
  if [ -n $cnf ]; then
      echo "found $cnf"
      mycli --defaults-file=$cnf
  else
      echo "term not found"
      usage
  fi 
else
    usage
fi

