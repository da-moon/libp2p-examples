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

	"github.com/libp2p/go-libp2p-net"
	json "github.com/multiformats/go-multicodec/json"

	multicodec "github.com/multiformats/go-multicodec"
)

// Payment protocol is used to send a transaction
// to another node
const paymentProtocol = "/payment/1.0.0"

// Ping protocol is used reply back a message
// to the node that sent out the transaction
const pingProtocol = "/ping/1.0.0"

// TransactionWrapper is a struct that is used to hold information relevant
// to a transaction such as <Sender> address, <Receiver> address and the <Amount>
// that is getting transferred
type TransactionWrapper struct {
	Sender   string
	Receiver string
	Amount   float64
}

// TransactionStream is a struct that is used to wrap a <net.Stream> stream.
// Since we are using <multicodec> to encode and decode streams and
// <bufio> to read to or write from streams, we put those in this struct so that
// we can easily 'carry' them with us and use them as we see fit.
type TransactionStream struct {
	stream  net.Stream
	writer  *bufio.Writer
	reader  *bufio.Reader
	encoder multicodec.Encoder
	decoder multicodec.Decoder
}

// WrapTransactionStream is the function that takes a stream and wrap it into a
// <TransactionStream> struct.We use JSON codec in this case
// ----------------------------------------------------------------------------
// <stream> is a parameter of <net.Stream> struct type that wee want to wrap into
// <TransactionStream> struct
// ----------------------------------------------------------------------------
// it returns a pointer to <TransactionStream> struct
func WrapTransactionStream(stream net.Stream) *TransactionStream {
	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)
	decoder := json.Multicodec(false).Decoder(reader)
	encoder := json.Multicodec(false).Encoder(writer)
	return &TransactionStream{
		stream:  stream,
		reader:  reader,
		writer:  writer,
		encoder: encoder,
		decoder: decoder,
	}
}

// sendTransaction is the function in which a <TransactionWrapper>
// struct is created and written to a <TransactionStream>
// ----------------------------------------------------------------------------
// <wrappedTransactionStream> is a receiver of pointer type to <TransactionStream>.It is
// the Wrapped stream that the Transsaction will be written to so that it
// can get transferred between nodes.
// ----------------------------------------------------------------------------
// <sender>,<receiver> and <amount> are parameters that are used to create the
// <TransactionWrapper> and write it to <wrappedTransactionStream>
// ----------------------------------------------------------------------------
// it returns an error if something goes wrong.
func (wrappedTransactionStream *TransactionStream) sendTransaction(sender string, receiver string, amount float64) error {
	// Use the functions input to create a <TransactionWrapper> called <tx>
	tx := &TransactionWrapper{
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
	}
	// use <wrappedTransactionStream.encoder> to encode <tx>
	err := wrappedTransactionStream.encoder.Encode(tx)
	// Write the transaction to the stream and since output is buffered with
	// <bufio> so <Flush> has to get called before exit.
	wrappedTransactionStream.writer.Flush()
	return err
}

// Payment is the main function that is used in payment
// protocol to send a transaction to another node.
// this function uses payment protocol to send the transaction
// and pingProtocol protocol to receives confirmation of transaction.
// ----------------------------------------------------------------------------
// <node> is a receiver of pointer type to <PeerNode>.
// <node> is the peer node that sends a transaction to another node
// ----------------------------------------------------------------------------
// <destination> is a parameter of string type that is the
// IPFS address of the node that receiving the transaction
// <amount> is a parameter of float64 type that represents the money
// getting transfered
func (node *PeerNode) Payment(destination string, amount float64) {
	// First, we add the peer node <destination> string points to
	// <node> local address book
	peerID, err := addAddressToPeerstore(node, destination)
	if err != nil {
		panic(err)
	}
	// <node> creates a news tream by calling  <NewStream>
	// function and passing background context, receiver's
	// <peerID> and <paymentProtocol> (<"/payment/1.0.0">)
	stream, err := node.NewStream(context.Background(), peerID, paymentProtocol)
	if err != nil {
		panic(err)
	}
	// use <WrapTransactionStream (stream net.Stream)> function to wrap
	// <stream> stream and save it in variable <wrappedTransactionStream>
	wrappedTransactionStream := WrapTransactionStream(stream)
	// it gets the <node> address and turns it into a default
	// IPFS address string and store it in variable <sender>
	sender := node.Addrs()[0].String()
	sender = fmt.Sprintf("%s/ipfs/%s", sender, node.ID().Pretty())
	// use <sendTransaction(sender string, receiver string, amount float64)>
	// on <wrappedTransactionStream> to send the transaction to receiver node.
	wrappedTransactionStream.sendTransaction(sender, destination, amount)
	// <node> creates a new stream  and storing it in variable <replyStream>
	// by calling  <NewStream> function and passing background context, receiver's
	// <peerID> and <pingProtocol> (<"/ping/1.0.0">)
	replyStream, err := node.NewStream(context.Background(), peerID, pingProtocol)
	if err != nil {
		panic(err)
	}
	// it reads back the <replyStream>. If the stream is sent
	// successfully,the stream is modified on the receiver node
	// so <ioutil.ReadAll()> is used to read back the
	// stream's content again confirm the receivers message that it recieved
	// the transaction.
	reply, err := ioutil.ReadAll(replyStream)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n %s => %s\n", reply, node.ID().String(), peerID)
	// close the stream
	stream.Close()

}

// decodeTransaction is used to decode a <*TransactionStream> into <*TransactionWrapper>
// ----------------------------------------------------------------------------
// <wrappedTransactionStream> is a receiver of pointer type to <TransactionStream>.It is
// the Wrapped stream that the Transsaction was written to so that it
// can get transferred between nodes and is ready to get decoded.
// ----------------------------------------------------------------------------
// it returns a pointer to <TransactionWrapper> which is the extracted value
// from the stream.
// It returns an erro in case something goes wrong.
func (wrappedTransactionStream *TransactionStream) decodeTransaction() (*TransactionWrapper, error) {
	var tx TransactionWrapper
	// use <wrappedTransactionStream.encoder> to decode and store to <tx>
	err := wrappedTransactionStream.decoder.Decode(&tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// PaymentProtocolMultiplexer : Multiplexes "/payment/1.0.0"
// and /ping/1.0.0 to a node and takes care of the way nodes behave when they
// receive a stream of payment protocol and ping protocol.
// It is called to initialize payment protocol before any other function
func (node *PeerNode) PaymentProtocolMultiplexer() {
	// <SetStreamHandler> function takes a string
	// (<pingProtocol>) and an anonymous function that
	// takes a <net.stream> struct and Multiplexes the string
	// to that anonymous functions code so whenever a stream
	// with the same string attached is received,that code
	// inside anonymous function is executed
	node.SetStreamHandler(pingProtocol, func(stream net.Stream) {
		// It prepares the <message> to get send back to the node that
		// sent the transaction
		message := fmt.Sprintf("\nTransaction Successful \t Thank You!\n")
		// it writes the <message> to stream in byte array
		// format so that the byte array is sent back to
		// the sender node
		_, err := stream.Write([]byte(message))
		if err != nil {
			panic(err)
		}
		stream.Close()

	})
	// <SetStreamHandler> function takes a string
	// (<paymentProtocol>) and an anonymous function that
	// takes a <net.stream> struct and Multiplexes the string
	// to that anonymous functions code so whenever a stream
	// with the same string attached is received by a node,that code
	// inside anonymous function is executed on the receiver node
	node.SetStreamHandler(paymentProtocol, func(stream net.Stream) {
		fmt.Println("Request Receiver : New connection intiated")
		// it uses <WrapTransactionStream (stream net.Stream)> function to wrap
		// <stream> stream and save it in variable <wrappedTransactionStream>
		wrappedTransactionStream := WrapTransactionStream(stream)
		// <decodeTransaction> is used to extract transaction from
		// stream and store it in variable <tx>
		tx, err := wrappedTransactionStream.decodeTransaction()
		if err != nil {
			// if transaction cannot be extracted, show the error and reset
			// the stream
			fmt.Println(err)
			wrappedTransactionStream.stream.Reset()
		} else {
			// if transaction is extracted, show the amount and
			// close the stream
			fmt.Printf("**********************************************************\n")
			fmt.Printf("Recieved Amount: %f\n", tx.Amount)
			fmt.Printf("**********************************************************\n")
			stream.Close()
		}

	})
	fmt.Printf("Payment Protocol 1.0.0 Multiplexd!\n")
}
