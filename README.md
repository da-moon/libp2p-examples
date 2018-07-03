# Libp2p-Examples
This repo uses [LibP2P library](https://github.com/libp2p/go-libp2p) to showcase some use case examples of LibP2P.
LibP2P has many functions and it can be quite challenging to simply know when to use what functions. In the code, I have commented every function extensively so that it can help people with their learning process so that they can use LibP2P in their own programs.  
Basic node functionalitis are in `basicNode.go` and all the others have corresponding names.
## Running
To run this, after compilation and running the compiled binary, type 'Help' to see the list of available options.

- [Heartbeat](#heartbeat)
- [Payment](#payment)
- [Sync](#sync)

## Heartbeat
In `heartbeat protocol` , I showcase the simplest use case of libp2p which is to have one node send one message to another node and the other node replies back with some message.

## Payment
in `payment protocol`, I showcase how two separate protocols can get multiplexed to a single node.
I also show how JSON-like messages can get transfered between nodes.
Here , I show the best practice on how to deal with data encoding, decoding and reading and writing to strings.
## Sync
in `sync protocol`, I show how one can use a coded such as `cbor` to request for file transfer from another node.
