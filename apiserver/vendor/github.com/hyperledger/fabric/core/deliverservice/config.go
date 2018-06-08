package deliverclient

import (
	"time"

	"github.com/spf13/viper"
)

var (
	configurationCached           = false
	reConnectMaxPeriod            = time.Second * 300
	reConnectMinPeriod            = time.Second * 5
	reConnectMinPeriodAttemptTime = 10
)

// cache the configuration
func cacheConfiguration() {
	max := viper.GetInt("peer.gossip.reConnectMaxPeriod")
	logger.Debugf("peer.gossip.reConnectMaxPeriod: %d", max)
	if max > 0 {
		reConnectMaxPeriod = time.Duration(max) * time.Second
	}
	min := viper.GetInt("peer.gossip.reConnectMinPeriod")
	logger.Debugf("peer.gossip.reConnectMinPeriod: %d", min)
	if min > 0 {
		reConnectMinPeriod = time.Duration(min) * time.Second
	}
	attemptTime := viper.GetInt("peer.gossip.reConnectMinPeriodAttemptTime")
	logger.Debugf("peer.gossip.reConnectMinPeriodAttemptTime: %d", attemptTime)
	if attemptTime > 0 {
		reConnectMinPeriodAttemptTime = attemptTime
	}
	configurationCached = true
}

func getReConnectMaxPeriod() time.Duration {
	if !configurationCached {
		cacheConfiguration()
	}
	return reConnectMaxPeriod
}

func getReConnectMinPeriod() time.Duration {
	if !configurationCached {
		cacheConfiguration()
	}
	return reConnectMinPeriod
}

func getReConnectMinPeriodAttemptTime() int {
	if !configurationCached {
		cacheConfiguration()
	}
	return reConnectMinPeriodAttemptTime
}
