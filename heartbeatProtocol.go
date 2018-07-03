/*The MIT License (MIT)
* Copyright (c) 2018 Damoon Azarpazhooh
* Permission is hereby granted, free of charge, to any person
* obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the
* Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute,
* sublicense, and/or sell copies of the Software, and to
* permit persons to whom the Software is furnished to do so,
* subject to the following conditions:
*
* The above copyright notice and this permission notice
* shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF
* ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
* PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
* OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
* OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
* OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"

	net "github.com/libp2p/go-libp2p-net"
)

// Heartbeat protocol is used to see if a node is still online.
// a node opens a stream to another node with a message and
// the other node would reply back with a message
const heartbeatprotocol = "/heartbeat/1.0.0"

// Heartbeat is the main function that is called for
// heartbeat protocol
// ----------------------------------------------------------
// <node> is a receiver of pointer type to <PeerNode>.
// <node> is the peer node that initiates a message to send
// to another node
// ----------------------------------------------------------
// <destination> is a parameter of string type that is the
// IPFS address of the node that is getting checked to see
// if it can receive messages or not.
func (node *PeerNode) Heartbeat(destination string) {
	// First, we add the peer node <destination> string points to
	// <node> local address book
	peerID, err := addAddressToPeerstore(node, destination)
	if err != nil {
		panic(err)
	}
	// <node> creates a new stream by calling  <NewStream>
	// function and passing background context, receiver's
	// <peerID> and <heartbeatprotocol> (<"/heartbeat/1.0.0">)
	stream, err := node.NewStream(context.Background(), peerID, heartbeatprotocol)
	if err != nil {
		panic(err)
	}
	// it would use <stream.Conn().LocalPeer()> to find the
	// peer id of the current node that is sending the message
	sender := stream.Conn().LocalPeer()
	// It creates the string <message> which is send
	// to stream receiver
	message := fmt.Sprintf(" %s is checking availablity\n", sender)
	// it writes the <message> to stream in byte array format
	// so that the byte array is sent to the receiver node
	_, err = stream.Write([]byte(message))
	if err != nil {
		panic(err)
	}
	// it reads back the stream. If the stream is sent
	// successfully,the stream is modified on the receiver node
	// so <ioutil.ReadAll()> is used to read back the
	// stream's content again.
	requestReceiver, err := ioutil.ReadAll(stream)
	if err != nil {
		panic(err)
	}
	// It shows the value of the stream after it was modified
	// on the receiver node.
	fmt.Printf("%s Reply: %s\n", peerID, requestReceiver)

}

// HeartbeatProtocolMultiplexer Multiplexes "/heartbeat/1.0.0"
// To a node and takes care of the way nodes behave when they
// receive a stream of heartbeat protocol
// It is called to initialize heartbeat protocol before any other function
// ----------------------------------------------------------------------------
// <node> is a receiver of pointer to struct type PeerNode
// and is the node that we multiplex "heartbeat protocol" to.
func (node *PeerNode) HeartbeatProtocolMultiplexer() {
	// <SetStreamHandler> function takes a string
	// (<heartbeatprotocol>) and an anonymous function that
	// takes a <net.stream> struct and Multiplexes the string
	// to that anonymous functions code so whenever a stream
	// with the same string attached is received by a node,that code
	// inside anonymous function is executed on the receiver node
	node.SetStreamHandler(heartbeatprotocol, func(stream net.Stream) {
		fmt.Println("Request Receiver : New connection intiated")
		// <bufio.NewReader(stream net.Stream)> is used to read
		// the data passed in the stream as buffer
		buf := bufio.NewReader(stream)
		// <ReadString()> is used t read the received string from
		// buffer <buf>
		str, err := buf.ReadString('\n')
		// check to make sure the stream doesn't have any issues.
		if err != nil {
			fmt.Println(err)
			stream.Reset()
		} else {
			// It shows the message it received.
			fmt.Printf("**********************************************************\n")
			fmt.Printf("Recieved Message: %s\n", str)
			fmt.Printf("**********************************************************\n")
			// it would use <stream.Conn().LocalPeer()> to
			// find the peer id of the current node that
			// received the message
			receiver := stream.Conn().LocalPeer()
			// It creates the string <message> which is used to
			// send back to stream sender
			message := fmt.Sprintf("%s recieved the message \t Node Is still online", receiver)
			// it writes the <message> to stream in byte array
			// format so that the byte array is sent back to
			// the sender node
			_, err = stream.Write([]byte(message))
			// close the stream
			stream.Close()
		}

	})
	fmt.Printf("Heartbeat Protocol 1.0.0 Multiplexd!\n")

}
