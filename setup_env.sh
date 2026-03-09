#!/bin/bash
# This script sources the .env file to export the environment variables.
# Call it ". ./source_env.sh"
# Or "source ./source_env.sh"

# Load the .env file
if [ -f ".env" ]; then
  while IFS='=' read -r key value; do
    # Skip lines that are comments or empty
    if [[ $key = \#* ]] || [[ -z $key ]]; then
      continue
    fi
    # Use printf to handle values with leading or trailing spaces
    value=$(printf '%s\n' "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    # Exporting each line as an environment variable
    export "$key=$value"
  done < ".env"
else
  echo "Missing .env file"
fi
