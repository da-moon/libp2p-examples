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
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/abiosoft/ishell"
)

func main() {
	shell()

}

func shell() {
	fmt.Printf("\nRun Help to see a list of options\n\n")
	myrand := random(1, 200)

	node := InitializePeer(myrand)

	shell := ishell.New()

	shell.AddCmd(&ishell.Cmd{
		Name: "protocols",
		Help: "list of implemented protocols",
		Func: func(c *ishell.Context) {
			choice := c.MultiChoice([]string{
				"Heartbeat",
				"Payment",
				"Sync",
			}, "Here is a list of implemented protocols.choose one!")
			switch choice {
			case 0:
				{
					shellHeartbeatOptions := ishell.New()
					node.HeartbeatProtocolMultiplexer()
					fmt.Println("Type 'Help' to see the list of available options for Heartbeat protocol ")
					shellHeartbeatOptions.AddCmd(&ishell.Cmd{
						Name: "connect",
						Help: "connect to a node",
						Func: func(c *ishell.Context) {
							c.Print("Server Address: ")
							receiverAddress := c.ReadLine()
							node.Heartbeat(receiverAddress)
						},
					})
					shellHeartbeatOptions.Run()
				}
			case 1:
				{
					shellPaymentOptions := ishell.New()
					node.PaymentProtocolMultiplexer()
					fmt.Println("Type 'Help' to see the list of available options for Payment protocol ")
					shellPaymentOptions.AddCmd(&ishell.Cmd{
						Name: "pay",
						Help: "pay a node",
						Func: func(c *ishell.Context) {
							c.Print("Receiver Address ")
							receiverAddress := c.ReadLine()
							c.Print("Amount? ")
							temp := c.ReadLine()
							amount, err := strconv.ParseFloat(temp, 64)
							if err != nil {
								panic(err)
							}
							node.Payment(receiverAddress, amount)
						},
					})
					shellPaymentOptions.Run()
				}
			case 2:
				{
					shellSyncOptions := ishell.New()
					node.SyncProtocolMultiplexer()
					fmt.Println("Type 'Help' to see the list of available options for Sync protocol ")
					shellSyncOptions.AddCmd(&ishell.Cmd{
						Name: "request",
						Help: "request files in /data folder of a target node",
						Func: func(c *ishell.Context) {
							c.Print("Target Node Address ")
							targetAddress := c.ReadLine()
							node.RequestSync(targetAddress)
						},
					})
					shellSyncOptions.Run()
				}
			default:
				fmt.Print("Try again")
			}
		},
	})
	shell.Run()
}
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
