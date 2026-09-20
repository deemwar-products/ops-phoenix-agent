#!/bin/sh
# Fail the container boot when API_UPSTREAM is missing.
set -e
if [ -z "${API_UPSTREAM}" ]; then
    echo "ERROR: API_UPSTREAM is not set." >&2
    echo "       Set it in deploy-ui.yml, e.g. http://agentproxy-api:8080" >&2
    exit 1
fi
echo "API_UPSTREAM=${API_UPSTREAM}"
