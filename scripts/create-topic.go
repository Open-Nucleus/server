// +build ignore

// Creates an HCS topic on Hedera testnet for Open Nucleus anchoring.
// Usage: go run scripts/create-topic.go
package main

import (
	"fmt"
	"os"

	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
)

func main() {
	// Use your ED25519 account (0.0.4560302)
	operatorID, err := hiero.AccountIDFromString("0.0.4560302")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad account ID: %v\n", err)
		os.Exit(1)
	}

	// Your DER-encoded private key — extract the raw key
	operatorKey, err := hiero.PrivateKeyFromStringDer("302e020100300506032b65700422042031e943e883bab5fd43071a90ab983c83e1b5cfd9981c6ac124512f012f0a5f69")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad private key: %v\n", err)
		os.Exit(1)
	}

	client := hiero.ClientForTestnet()
	client.SetOperator(operatorID, operatorKey)
	defer client.Close()

	// Create anchor topic
	txResponse, err := hiero.NewTopicCreateTransaction().
		SetTopicMemo("Open Nucleus — Clinical Data Integrity Anchoring").
		SetAdminKey(operatorKey.PublicKey()).
		SetSubmitKey(operatorKey.PublicKey()).
		Execute(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create topic failed: %v\n", err)
		os.Exit(1)
	}

	receipt, err := txResponse.GetReceipt(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get receipt failed: %v\n", err)
		os.Exit(1)
	}

	topicID := receipt.TopicID
	fmt.Printf("Anchor Topic created: %s\n", topicID)

	// Create DID topic
	txResponse2, err := hiero.NewTopicCreateTransaction().
		SetTopicMemo("Open Nucleus — DID Document Registry").
		SetAdminKey(operatorKey.PublicKey()).
		SetSubmitKey(operatorKey.PublicKey()).
		Execute(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create DID topic failed: %v\n", err)
		os.Exit(1)
	}

	receipt2, err := txResponse2.GetReceipt(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get DID receipt failed: %v\n", err)
		os.Exit(1)
	}

	didTopicID := receipt2.TopicID
	fmt.Printf("DID Topic created:    %s\n", didTopicID)

	fmt.Println()
	fmt.Println("Add these to your config.yaml:")
	fmt.Printf("  topic_id: \"%s\"\n", topicID)
	fmt.Printf("  did_topic_id: \"%s\"\n", didTopicID)
}
