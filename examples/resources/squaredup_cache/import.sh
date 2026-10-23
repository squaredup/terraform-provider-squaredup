# Caches are imported by their cache ID. List your caches with the API to find
# it. Each row shows the cache ID, cache name, data stream and table name. The
# table name is unique and shown in Settings > Data > Caches > Edit cache.
curl -s "https://api.squaredup.com/api/smartcache?apiKey=$SQUAREDUP_API_KEY" |
  jq -r '.[] | [.id, (.displayName // .name), .steps[0].dataStreamName, .steps[0].smartCacheMapping.tableName] | @tsv'

# Then import the cache with its ID
terraform import squaredup_cache.example smartcachedefinition-123

# After importing, add the cache's settings to your configuration until
# terraform plan shows no changes. A Gremlin query or collection scope isn't
# imported, so add it yourself.
