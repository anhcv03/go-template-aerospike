package db

import (
	"fmt"
	"regexp"
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

var (
	infoErrRegexp = regexp.MustCompile(`(?i)(fail|error)((:|=)(?P<code>[0-9]+))?((:|=)(?P<msg>.+))?`)
	loadPctRegexp = regexp.MustCompile(`load_pct=(\d+)`)
)

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

	command := "namespace/" + ad.DbConfig.Namespace
	info, err := nodes[0].RequestInfo(as.NewInfoPolicy(), command)
	if err != nil {
		return err
	}

	value, ok := info[command]
	if !ok || value == "" {
		return fmt.Errorf("namespace %s not found", ad.DbConfig.Namespace)
	}

	if strings.HasPrefix(strings.ToUpper(value), "ERROR") {
		return fmt.Errorf("namespace %s not found: %s", ad.DbConfig.Namespace, value)
	}

	return nil
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
	nodes := client.GetNodes()
	if len(nodes) == 0 {
		return fmt.Errorf("aerospike cluster is empty")
	}

	command := buildCreateSecondaryIndexCommand(namespace, setName, indexName, binName, indexType)
	response, err := requestInfoValue(nodes[0], command)
	if err != nil {
		return err
	}

	if !strings.EqualFold(response, "ok") && !infoResponseMatchesCode(response, astypes.INDEX_FOUND) {
		return fmt.Errorf("create index %s failed: %s", indexName, response)
	}

	return waitForSecondaryIndex(nodes, namespace, indexName, 30*time.Second)
}

func buildCreateSecondaryIndexCommand(namespace, setName, indexName, binName string, indexType as.IndexType) string {
	return fmt.Sprintf(
		"sindex-create:ns=%s;set=%s;indexname=%s;indexdata=%s,%s",
		namespace,
		setName,
		indexName,
		binName,
		indexType,
	)
}

func requestInfoValue(node *as.Node, command string) (string, error) {
	info, err := node.RequestInfo(as.NewInfoPolicy(), command)
	if err != nil {
		return "", err
	}

	response, ok := info[command]
	if !ok || response == "" {
		return "", fmt.Errorf("info command %q returned empty response", command)
	}

	return response, nil
}

func waitForSecondaryIndex(nodes []*as.Node, namespace, indexName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		ready, err := secondaryIndexReady(nodes, namespace, indexName)
		if err == nil && ready {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return err
			}
			return fmt.Errorf("wait for index %s timed out", indexName)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func secondaryIndexReady(nodes []*as.Node, namespace, indexName string) (bool, error) {
	for _, node := range nodes {
		ready, err := secondaryIndexReadyOnNode(node, namespace, indexName)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
	}

	return true, nil
}

func secondaryIndexReadyOnNode(node *as.Node, namespace, indexName string) (bool, error) {
	commands := []string{
		fmt.Sprintf("sindex-stat:namespace=%s;indexname=%s", namespace, indexName),
		fmt.Sprintf("sindex/%s/%s", namespace, indexName),
	}

	var lastErr error

	for _, command := range commands {
		response, err := requestInfoValue(node, command)
		if err != nil {
			lastErr = err
			continue
		}

		ready, err := secondaryIndexReadyResponse(response)
		if err == nil {
			return ready, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("index %s status unavailable", indexName)
	}

	return false, lastErr
}

func secondaryIndexReadyResponse(response string) (bool, error) {
	match := loadPctRegexp.FindStringSubmatch(response)
	if len(match) == 2 {
		pct, err := strconv.Atoi(match[1])
		if err != nil {
			return false, fmt.Errorf("invalid load_pct value %q", match[1])
		}
		return pct >= 100, nil
	}

	if infoResponseMatchesCode(response, astypes.INDEX_NOTFOUND) || infoResponseMatchesCode(response, astypes.INDEX_NOTREADABLE) {
		return false, nil
	}

	upper := strings.ToUpper(response)
	if strings.Contains(upper, "FAIL") || strings.Contains(upper, "ERROR") {
		return false, fmt.Errorf("index status failed: %s", response)
	}

	return false, nil
}

func infoResponseMatchesCode(response string, code astypes.ResultCode) bool {
	match := infoErrRegexp.FindStringSubmatch(response)
	if len(match) > 0 {
		for i, name := range infoErrRegexp.SubexpNames() {
			if name != "code" || match[i] == "" {
				continue
			}
			value, err := strconv.Atoi(match[i])
			if err == nil && astypes.ResultCode(value) == code {
				return true
			}
		}
	}

	return strings.Contains(strings.ToUpper(response), astypes.ResultCodeToString(code))
}
