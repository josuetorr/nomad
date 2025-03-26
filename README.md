# Nomad
A decentralizied search engine

# Brainstorm
the index should map from the document url to the terms and term frequencies

index[url] => termfreq[term] => freq

for now, we'll save the index on disk. Maybe later we'll find a way to speed up the process since the index will become huge

# TechStack
A brief overview of nomad's techstack
* Language -> golang
* Crawler ->  colly
* Indexer -> custome indexing
* Query routing -> kademlia / websockets
* Storage -> rocksDB / ipfs
* P2P Network -> libp2p / kademlia dht
