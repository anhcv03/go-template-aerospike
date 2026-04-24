package common

import (
	as "github.com/aerospike/aerospike-client-go/v8"
	"fmt"
)

func ProbeReadiness(dbClient *as.Client) error {
	nodes := dbClient.GetNodes()
	if len(nodes) == 0 {
		return fmt.Errorf("aerospike cluster is empty")
	}

	_, err := nodes[0].RequestInfo(as.NewInfoPolicy(), "build")
	return err
}

