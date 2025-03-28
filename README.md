# Nomad
A decentralizied search engine

# Brainstorm
Each node will be in charge of 3 tasks: 
* crawling
* indexing
* searching

crawling and indexing will be performed concurrently using goroutines and channels. When a page has been crawled, we will index said page. We will store
the page itself and it's index values


# TechStack
A brief overview of nomad's techstack
* Language -> golang
* Crawler ->  colly
* Indexer -> custome indexing
* Query routing -> kademlia / websockets
* Storage -> rocksDB / ipfs
* P2P Network -> libp2p / kademlia dht
