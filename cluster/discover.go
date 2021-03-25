package cluster

/*
this should be ssdp stuff.

For sanity, we should only broadcast search messages when unclustered, and we should only listen for search messages of our type.
That way, we can hopefully reduce broadcast noise.

Should still accept a seed node in lieu of discovery, but for now assume discovery.
*/
