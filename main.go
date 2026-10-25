package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	client, err := NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx := context.Background()
	created, err := client.CreateKey(ctx, "temporary-rotation-key")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// The API returns the new plaintext key once. Store it now; it cannot be retrieved a second time.
	var createdObj struct {
		ID string `json:"key_id"`
	}
	_ = json.Unmarshal(created, &createdObj)
	fmt.Printf("temporary key created: %s\n", string(created))
	if createdObj.ID == "" {
		fmt.Fprintln(os.Stderr, "create response did not include a key id")
		os.Exit(1)
	}
	if err := client.RotateKey(ctx, createdObj.ID, 24); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logs, err := client.SearchLogs(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("old-key deployment evidence: %s\n", string(logs))
}
