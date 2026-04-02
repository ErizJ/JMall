#!/bin/sh
# Replace DB and Redis addresses in the yaml config with values from environment variables.
# This avoids baking secrets into the image.

CONFIG="/app/etc/${SERVICE}-api.yaml"

if [ -n "$DB_SOURCE" ]; then
  sed -i "s|DataSource:.*|DataSource: ${DB_SOURCE}|" "$CONFIG"
fi

if [ -n "$REDIS_ADDR" ]; then
  sed -i "s|Addr:.*|Addr: ${REDIS_ADDR}|" "$CONFIG"
fi

exec /app/service -f "$CONFIG"
