# Caches are imported by their cache ID. The ID isn't shown in SquaredUp, and
# cache names don't have to be unique, so list your caches with the API to find
# it. Use https://eu.api.squaredup.com for an EU organization. Each row shows the
# cache ID, cache name, data stream and table name. The table name is unique in
# your organization and is shown in Settings > Data > Caches > Edit cache, so use
# it to pick the right cache.
curl -s "https://api.squaredup.com/api/smartcache?apiKey=$SQUAREDUP_API_KEY" |
  jq -r '.[] | [.id, (.displayName // .name), .steps[0].dataStreamName, .steps[0].smartCacheMapping.tableName] | @tsv'

# Then import the cache with its ID
terraform import squaredup_cache.example smartcachedefinition-123

# After importing, add the cache's settings to your configuration until
# terraform plan shows no changes. A scope set as a Gremlin query or a
# collection isn't imported, so add it to your configuration yourself.
