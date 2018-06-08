package couchdb

import (
	"fmt"
	pkg_config "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/config"
	"github.com/hyperledger/fabric/core/ledger/util/couchdb"
	"github.com/spf13/viper"
	"time"
	"sync"
)

var (
	couchDBClients []*couchdb.CouchDatabase
	m                   sync.Mutex
)

func GetDBClients() ([]*couchdb.CouchDatabase, error) {
	if couchDBClients == nil {
		m.Lock()
		defer m.Unlock()
		if couchDBClients == nil {
			var err error
			couchDBClients, err = innerCouchDBClients()
			if err != nil {
				return nil, fmt.Errorf("GetCouchDBClients failed:%s", err.Error())
			}
		}
	}
	return couchDBClients, nil
}

func innerCouchDBClients() ([]*couchdb.CouchDatabase, error) {
	conf, err := pkg_config.GetCouchDBConfig()
	if err != nil {
		return nil, fmt.Errorf("GetCouchDBConfig failed:%s", err.Error())
	}
	chainID := viper.GetString("chaincode.id.chainID")
	var clients []*couchdb.CouchDatabase
	for _, info := range conf.CouchDbs {
		cli, err := GetCouchDatabase(info.CouchDBAddress, info.Username, info.Password, chainID, conf.MaxRetries, conf.MaxRetriesOnStartup, conf.RequestTimeout)
			if err != nil {
			return nil, fmt.Errorf("GetCouchDatabase failed:%s", err.Error())
		}
		clients = append(clients, cli)
	}
	return clients, nil
}

func GetCouchDatabase(addr, name, pwd, dbName string, maxRetries, maxRetriesOnStartup int, requestTimeout string) (*couchdb.CouchDatabase, error) {
	reqTimeout, err := time.ParseDuration(requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("ParseDuration failed:%s", err.Error())
	}
	couchInstance, err := couchdb.CreateCouchInstance(addr, name, pwd,
		maxRetries, maxRetriesOnStartup, reqTimeout)
	if err != nil {
		return nil, err
	}
	// CreateCouchDatabase creates a CouchDB database object, as well as the underlying database if it does not exist
	db, err := couchdb.CreateCouchDatabase(*couchInstance, dbName)
	if err != nil {
		return nil, err
	}
	return db, nil
}

