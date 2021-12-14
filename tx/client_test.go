package tx

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"testing"
)

func TestClientBatchRequest(t *testing.T) {
	client := NewWeb3Client("https://bsc-dataseed1.defibit.io/")
	fq := FilterQuery{
		Addresses: []common.Address{common.HexToAddress("0x7b4452dd6c38597fa9364ac8905c27ea44425832")},
	}

	flowable := client.EthLogFlowable(fq, 1000)
	for item := range flowable {
		ethLog := item.(types.Log)
		fmt.Println(ethLog)
	}

}
