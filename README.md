# Celesteum

Simple project that pulls all transaction hashes from latest block on ethereum testnet (Goerli network) and posts the data on to a Celestia light node.

Pass in `namespaceId` as a command line flag, whichis the namespace id of the Data Availability layer the messages will be submitted to. Also create a `.env` file and add values for `CELESTIA_NODE_URL` and `API_KEY` which is your alchemy API KEY.