package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	as "github.com/aerospike/aerospike-client-go/v8"
	aslogger "github.com/aerospike/aerospike-client-go/v8/logger"
	astypes "github.com/aerospike/aerospike-client-go/v8/types"
	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"

	"github.com/rs/zerolog/log"
)

type AerospikeDB struct {
	DbConfig *config.AerospikeConfig
}

func NewAerospikeConnection(dbCfg config.AerospikeConfig) *AerospikeDB {
	return &AerospikeDB{
		DbConfig: &dbCfg,
	}
}

func (ad *AerospikeDB) ConnectDatabase() *as.Client {
	clientPolicy := as.NewClientPolicy()
	clientPolicy.Timeout = time.Duration(ad.DbConfig.ConnectTimeoutMs) * time.Millisecond
	clientPolicy.ConnectionQueueSize = ad.DbConfig.ConnectionQueueSize
	clientPolicy.LimitConnectionsToQueueSize = true
	clientPolicy.MinConnectionsPerNode = ad.DbConfig.MinConnectionsPerNode
	clientPolicy.IdleTimeout = time.Duration(ad.DbConfig.IdleTimeoutMs) * time.Millisecond
	clientPolicy.TendInterval = time.Duration(ad.DbConfig.TendIntervalMs) * time.Millisecond
	clientPolicy.UseServicesAlternate = ad.DbConfig.UseServicesAlternate
	clientPolicy.User = ad.DbConfig.User
	clientPolicy.Password = ad.DbConfig.Pass

	aslogger.Logger.SetLevel(aslogger.WARNING)

	hosts, err := buildHosts(*ad.DbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid aerospike host configuration")
		return nil
	}

	client, err := as.NewClientWithPolicyAndHost(clientPolicy, hosts...)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to aerospike")
		return nil
	}

	if err := ad.ensureNamespaceExists(client); err != nil {
		client.Close()
		log.Fatal().Err(err).Msg("aerospike namespace validation failed")
		return nil
	}

	if ad.DbConfig.WarmUpConnections > 0 {
		if warmed, warmErr := client.WarmUp(ad.DbConfig.WarmUpConnections); warmErr != nil {
			log.Warn().Err(warmErr).Int("warmed_connections", warmed).Msg("aerospike warm up incomplete")
		}
	}

	return client
}

func (ad *AerospikeDB) ensureNamespaceExists(client *as.Client) error {
	nodes := client.GetNodes()
	if len(nodes) == 0 {
		return fmt.Errorf("aerospike cluster is empty")
	}

	info, err := nodes[0].RequestInfo(as.NewInfoPolicy(), "namespaces")
	if err != nil {
		return err
	}

	nsList, ok := info["namespaces"]
	if !ok || nsList == "" {
		return fmt.Errorf("failed to retrieve namespaces")
	}

	namespaces := strings.Split(nsList, ";")
	for _, ns := range namespaces {
		if ns == ad.DbConfig.Namespace {
			return nil
		}
	}

	return fmt.Errorf("namespace %s not found", ad.DbConfig.Namespace)
}

func buildHosts(cfg config.AerospikeConfig) ([]*as.Host, error) {
	if len(cfg.Hosts) == 0 {
		return []*as.Host{as.NewHost(cfg.Host, cfg.Port)}, nil
	}

	hosts := make([]*as.Host, 0, len(cfg.Hosts))
	for _, rawHost := range cfg.Hosts {
		rawHost = strings.TrimSpace(rawHost)
		if rawHost == "" {
			continue
		}

		host := rawHost
		port := cfg.Port

		parts := strings.Split(rawHost, ":")
		if len(parts) > 1 {
			host = strings.Join(parts[:len(parts)-1], ":")
			parsedPort, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return nil, fmt.Errorf("parse aerospike host port %q: %w", rawHost, err)
			}
			port = parsedPort
		}

		hosts = append(hosts, as.NewHost(host, port))
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no aerospike hosts configured")
	}

	return hosts, nil
}

func EnsureSecondaryIndex(client *as.Client, namespace, setName, indexName, binName string, indexType as.IndexType) error {
	task, err := client.CreateIndex(nil, namespace, setName, indexName, binName, indexType)
	if err != nil {
		if err.Matches(astypes.INDEX_FOUND) {
			return nil
		}
		return err
	}

	if task == nil {
		return nil
	}

	return <-task.OnComplete()
}
