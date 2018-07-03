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
	"os"

	"github.com/libp2p/go-libp2p-net"
	"github.com/mholt/archiver"
	"github.com/multiformats/go-multicodec"

	cbor "github.com/multiformats/go-multicodec/cbor"
)

// Sync protocol is used to sync files in a directory and send them
// to another node.
const syncProtocol = "/sync/1.0.0"

// DataStream is a struct that is used to wrap a <net.Stream> stream.
// Since we are using <multicodec> to encode and decode streams and
// <bufio> to read to or write from streams, we put those in this struct so that
// we can easily 'carry' them with us and use them as we see fit.
type DataStream struct {
	stream  net.Stream
	encoder multicodec.Encoder
	decoder multicodec.Decoder
	writer  *bufio.Writer
	reader  *bufio.Reader
}

// WrapDataStream is the function that takes a stream and wrap it into a
// <DataStream> struct.We use cbor codec in this case
// ----------------------------------------------------------------------------
// <stream> is a parameter of <net.Stream> struct type that wee want to wrap into
// <DataStream> struct
// ----------------------------------------------------------------------------
// it returns a pointer to <DataStream> struct
func WrapDataStream(stream net.Stream) *DataStream {
	var result DataStream
	result.reader = bufio.NewReader(stream)
	result.writer = bufio.NewWriter(stream)
	result.decoder = cbor.Multicodec().Decoder(result.reader)
	result.encoder = cbor.Multicodec().Encoder(result.writer)
	return &result
}

// decodeTransfer is used to decode a <*DataStream> and save the files
// encoded onto the disk
// ----------------------------------------------------------------------------
// <wrappedDataStream> is a receiver of pointer type to <DataStream>.It is
// the Wrapped stream that the file was written to so that it
// can get transferred between nodes and is ready to get decoded.
// ----------------------------------------------------------------------------
// It returns an erro in case something goes wrong.
func (wrappedDataStream *DataStream) decodeTransfer() error {
	// initialize variable <file> as an array of bytes
	var file []byte
	// use <wrappedDataStream.encoder> to decode and store to <file>
	err := wrappedDataStream.decoder.Decode(&file)
	if err != nil {
		return err
	}
	// use <ioutil.WriteFile> to write <file> byte array to disk as the zip file
	// it originated from.
	err = ioutil.WriteFile("data.zip", file, 0644)
	// if there is an error, remove the zip file and return an error
	if err != nil {
		os.Remove("data.zip")
		return (err)
	}
	// use <archiver> package to unzip the zip file that was transferred in
	// the stream and was stored on disk.
	err = archiver.Zip.Open("data.zip", "")
	// If there is an error, remove the zip file and < /data > directory in
	// which contents of the file were supposed to get unzipped to.
	if err != nil {
		os.Remove("data.zip")
		os.Remove("data")
		return err
	}
	// if there are no issues, just remove the zip file from hard drive.
	os.Remove("data.zip")
	return nil
}

// RequestSync : the main function that is used in sync
// protocol to request files from another node that has the files..
// ----------------------------------------------------------------------------
// <node> is a receiver of pointer type to <PeerNode>.
// <node> is the peer node that requests the files
// ----------------------------------------------------------------------------
// <wrappedTransactionStream> is a receiver of pointer type to <TransactionStream>.It is
// the Wrapped stream that the Transsaction will be written to so that it
// can get transferred between nodes.
// <bootstrapNodeAddress> is a parameter of string type that is the
// IPFS address of the node that receiving the request to transfer the files
func (node *PeerNode) RequestSync(bootstrapNodeAddress string) {
	// First, we add the peer node <bootstrapNodeAddress> string points to
	// <node> local address book

	peerID, err := addAddressToPeerstore(node, bootstrapNodeAddress)
	if err != nil {
		panic(err)
	}
	// <node> creates a news tream by calling  <NewStream>
	// function and passing background context, receiver's
	// <peerID> and <paymentProtocol> (<"/sync/1.0.0">)
	stream, err := node.NewStream(context.Background(), peerID, syncProtocol)
	if err != nil {
		panic(err)
	}
	// we make sure the stream gets closed at the end of the function call by
	// using defer keyword before <stream.Close()>.
	defer stream.Close()
	// use <WrapDataStream (stream net.Stream)> function to wrap
	// <stream> stream and save it in variable <wrappedDataStream>
	wrappedDataStream := WrapDataStream(stream)
	// Call <decodeTransfer()> to save the received Zip file on disk and
	// extract it.
	err = wrappedDataStream.decodeTransfer()
	if err != nil {
		panic(err)
	}

}

// encodeFile is the function in which all the files in </data>
// are zipped and transferredover a strem.
// ----------------------------------------------------------------------------
// <wrappedDataStream> is a receiver of pointer type to <DataStream>.It is
// the Wrapped stream that the a zip file will be written to so that it
// can get transferred between nodes.
// ----------------------------------------------------------------------------
// it returns an error if something goes wrong.
func (wrappedDataStream *DataStream) encodeFile() error {
	// use <archiver> package to zip the files in < /data > directory
	// so that they can easily get transferred over the stream
	err := archiver.Zip.Make("data.zip", []string{"data"})
	if err == nil {
		// use <ioutil.ReadFile> to read data.zip from disk to byte array <data>
		data, err := ioutil.ReadFile("data.zip")
		if err != nil {
			panic(err)
		}
		// use <wrappedDataStream.encoder> to encode <data>
		err = wrappedDataStream.encoder.Encode(data)
		// Write the transaction to the stream and since output is buffered with
		// <bufio> so <Flush> has to get called before exit.
		wrappedDataStream.writer.Flush()
		// Remove data.zip from disk
		os.Remove("data.zip")

		return err
	}

	return err

}

// SyncProtocolMultiplexer Multiplexes "/sync/1.0.0" to a node and takes care
// of the way nodes behave when they receive a stream of sync protocol.
// It is called to initialize payment protocol before any other function
func (node *PeerNode) SyncProtocolMultiplexer() {
	// <SetStreamHandler> function takes a string
	// (<syncProtocol>) and an anonymous function that
	// takes a <net.stream> struct and Multiplexes the string
	// to that anonymous functions code so whenever a stream
	// with the same string attached is received by a node,that code
	// inside anonymous function is executed on the receiver node
	node.SetStreamHandler(syncProtocol, func(stream net.Stream) {
		fmt.Println("Sync intiated!")
		// it uses <WrapDataStream (stream net.Stream)> function to wrap
		// <stream> stream and save it in variable <wrappedDataStream>
		wrappedDataStream := WrapDataStream(stream)
		// use <encodeFile()> function to zip the files in < /data > directory
		// and write them to the stream.
		err := wrappedDataStream.encodeFile()

		if err != nil {
			fmt.Println(err)
			stream.Reset()

		} else {
			stream.Close()
		}
	})
	fmt.Printf("Sync Protocol 1.0.0 Multiplexd!\n")
}
