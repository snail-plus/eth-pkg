package tx

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	notFoundErrorStr = "filter not found"
)

type FilterQuery struct {
	BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash
	FromBlock rpc.BlockNumber  // beginning of the queried range, nil means genesis block
	ToBlock   rpc.BlockNumber  // end of the range, nil means latest block
	Addresses []common.Address // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]common.Hash
}

type Filter interface {
	GetFilterId() (string, error)
	Type() reflect.Type
}

type BaseFilter struct {
	FilterId     string
	Filter       Filter
	FilterQuery  FilterQuery
	rpcClient    *rpc.Client
	pullInterval int64
	LogChan      chan interface{}
}

func (b BaseFilter) Run(pullInterval int64) {
	b.pullInterval = pullInterval
	filterId, err := b.Filter.GetFilterId()
	if err != nil {
		log.Printf("get filterId error: %s", err.Error())
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(pullInterval) * time.Millisecond)
		defer func() {
			ticker.Stop()
			b.ReInstall()
		}()

		for range ticker.C {
			var err error
			typeName := b.Filter.Type().Kind()
			switch typeName {
			case reflect.String:
				var hashArr []string
				err = b.rpcClient.Call(&hashArr, "eth_getFilterChanges", filterId)
				if err == nil {
					for _, item := range hashArr {
						b.LogChan <- item
					}
				}

			case reflect.Struct:
				var ethLogArr []types.Log
				err = b.rpcClient.Call(&ethLogArr, "eth_getFilterChanges", filterId)
				if err == nil {
					for _, item := range ethLogArr {
						b.LogChan <- item
					}
				}
			}

			if err != nil {
				log.Printf("eth_getFilterChanges error: %s", err.Error())
				if strings.Contains(err.Error(), notFoundErrorStr) {
					return
				}
			}
		}
	}()

}

func (b BaseFilter) ReInstall() {
	b.Run(b.pullInterval)
}

type PendingTransactionFilter struct {
	*BaseFilter
}

func (f *PendingTransactionFilter) GetFilterId() (string, error) {
	var filterID string
	err := f.rpcClient.Call(&filterID, "eth_newPendingTransactionFilter")
	return filterID, err
}

func (f *PendingTransactionFilter) Type() reflect.Type {
	return reflect.TypeOf("")
}

func NewPendingTransactionFilter(rpcClient *rpc.Client) *PendingTransactionFilter {
	p := &PendingTransactionFilter{}
	p.BaseFilter = &BaseFilter{
		Filter:    p,
		rpcClient: rpcClient,
		LogChan:   make(chan interface{}, 500),
	}
	return p
}

type LogFilter struct {
	*BaseFilter
}

func (f *LogFilter) GetFilterId() (string, error) {
	var filterID string
	err := f.rpcClient.Call(&filterID, "eth_newFilter", f.FilterQuery)
	return filterID, err
}

func (f *LogFilter) Type() reflect.Type {
	return reflect.TypeOf(types.Log{})
}

func NewLogFilterFilter(rpcClient *rpc.Client, filterQuery FilterQuery) *LogFilter {
	l := &LogFilter{}
	l.BaseFilter = &BaseFilter{
		Filter:      l,
		FilterQuery: filterQuery,
		rpcClient:   rpcClient,
		LogChan:     make(chan interface{}, 5000),
	}
	return l
}
