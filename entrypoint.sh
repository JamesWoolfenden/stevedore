#!/bin/bash

# Leverage the default env variables as described in:
# https://docs.github.com/en/actions/reference/environment-variables#default-environment-variables
if [[ $GITHUB_ACTIONS != "true" ]]
then
  /usr/bin/stevedore "$@"
  exit $?
fi

flags=""

echo "running command:"
echo stevedore label -f "$INPUT_FILE" "$flags"

/usr/bin/stevedore label -f "$INPUT_FILE" "$flags"
export stevedore_EXIT_CODE=$?
