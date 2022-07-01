package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	channelName   = "mychannel" // 通道名称
	username      = "Admin"     // 用户
	chainCodeName = "basic"     // 链码名称
)

func main() {
	sdk, err := fabsdk.New(config.FromFile("config.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	defer sdk.Close()

	ctx := sdk.ChannelContext(channelName, fabsdk.WithUser(username))
	client, err := channel.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Query(channel.Request{ChaincodeID: chainCodeName, Fcn: "query", Args: [][]byte{[]byte("a")}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.TransactionID)
	fmt.Println(result.TxValidationCode)
	fmt.Println(result.ChaincodeStatus)
	fmt.Println(string(result.Payload))

	result, _= client.Query(channel.Request{ChaincodeID: chainCodeName, Fcn: "query", Args: [][]byte{[]byte("b")}})
	fmt.Println(string(result.Payload))

	result, err = client.Execute(channel.Request{ChaincodeID: chainCodeName, Fcn: "invoke", Args: [][]byte{[]byte("a"), []byte("b"), []byte("10")}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.TransactionID)
	fmt.Println(result.TxValidationCode)
	fmt.Println(result.ChaincodeStatus)
	fmt.Println(string(result.Payload))
}
