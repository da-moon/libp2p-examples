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
	"context"
	"crypto/rand"
	"fmt"
	"io"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// PeerNode is a struct that is used as a wrapper for host.Host structs
// so that using receiver style function calls becomes possible
type PeerNode struct {
	host.Host
}

// InitializePeer function is the starting point for any P2P application.
// This function creates a node
// ----------------------------------------------------------------------------
// <SourcePort> is an integer parameter that indicates which port the process
// sends or receives packets FROM
// ----------------------------------------------------------------------------
// It returns a pointer to a *PeerNode struct type so that it can be used as a
// receiver type.
func InitializePeer(sourcePort int) *PeerNode {
	var r io.Reader
	r = rand.Reader
	var result *PeerNode
	// Generate a 2018 private key with RSA
	// You chould be able to substitue it with other cryptography functions such as
	//  elliptic curves.
	privateKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	// Generate a IP4 TCp multi address and point it to 0.0.0.0 as a way to say that
	// It accepts all connections.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", sourcePort))
	// Use the current context, generated private key as Identity and attach
	// the generated multi address the links to 0.0.0.0 to create a new peer Node
	// 0.0.0.0 tells our host to accept all addresses
	// <Context package> 		https://golang.org/pkg/context/
	node, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(privateKey),
	)
	if err != nil {
		panic(err)
	}
	// Show the created node properties on display and return a pointer to it.
	fmt.Printf("Node PeerID:\t%s\n", node.ID())
	fmt.Printf("\n%s/ipfs/%s\n", node.Addrs()[0].String(), node.ID().Pretty())
	result = &PeerNode{node}

	return result
}

// addAddressToPeerstore function adds an address to a node's address book
// ----------------------------------------------------------------------------
// <node> is a parameter of struct type host.Host is the node we want to add a
// peer addess to it's address book
// <address> is a parameter of string type that is of IPFS address type and
// it is the address we want to add to address book.
// ----------------------------------------------------------------------------
// It returns a peer.ID which is the decoded peer.ID of <address>
// It also returs an error to be used in case it is needed
func addAddressToPeerstore(node host.Host, address string) (peer.ID, error) {
	var result peer.ID
	// Encapsulate the given string <address> into a multiaddress variable called
	// <ipfsAddress>
	ipfsAddress, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return result, err
	}
	// use <ValueForProtocol(multiaddr.[...])> function to turn back the <ipfsAddress>
	// multiaddress into it's full base 58 representation <peerIDString>
	peerIDString, err := ipfsAddress.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return result, err
	}
	// We generate a full address form by creating a new multiaddress and
	// passing the bsse 58 encoded address to generate target peer address
	targetPeerAddress, _ := multiaddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peerIDString))

	// decode <peerIDString> base 58 representation into <decodedPeerID>
	decodedPeerID, err := peer.IDB58Decode(peerIDString)
	if err != nil {
		return result, err
	}
	result = decodedPeerID
	// We use decapsulate target peer address multiwrapper, and decoded peer
	// add use the following function to add them to address book
	targetAddress := ipfsAddress.Decapsulate(targetPeerAddress)
	node.Peerstore().AddAddr(decodedPeerID, targetAddress, peerstore.PermanentAddrTTL)

	return result, err
}

// IpfsAddressToPeerID turns a IPFS address into its corresponding peer ID
// ----------------------------------------------------------------------------
// <address> is a parameter of string type that is of IPFS address type and
// it is the address we want to get the corresponding <peer.ID>
// ----------------------------------------------------------------------------
// <peer.ID> is the result we are looking for
// it returns an error in case there is an issue in converting the string into peer.ID
func IpfsAddressToPeerID(address string) (peer.ID, error) {
	var result peer.ID
	// Encapsulate the given string <address> into a multiaddress variable called
	// <ipfsAddress>
	ipfsAddress, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return result, err
	}
	// use <ValueForProtocol(multiaddr.[...])> function to turn back the <ipfsAddress>
	// multiaddress into it's full base 58 representation <peerIDString>.
	peerIDString, err := ipfsAddress.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return result, err
	}
	// decode <peerIDString> base 58 representation into <decodedPeerID>
	decodedPeerID, err := peer.IDB58Decode(peerIDString)
	if err != nil {
		return result, err
	}
	// return <decodedPeerID> or error
	result = decodedPeerID
	return result, err
}
