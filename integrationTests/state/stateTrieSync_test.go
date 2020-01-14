package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()
	defer func() {
		_ = nRequester.Node.Stop()
		_ = nResolver.Node.Stop()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	stateTrie := nResolver.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	_ = stateTrie.Update([]byte("doe"), []byte("reindeer"))
	_ = stateTrie.Update([]byte("dog"), []byte("puppy"))
	_ = stateTrie.Update([]byte("dogglesworth"), []byte("cat"))
	_ = stateTrie.Commit()
	rootHash, _ := stateTrie.Root()

	requesterStateTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterStateTrie.Root()
	trieNodesResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.AccountTrieNodesTopic)

	waitTime := 5 * time.Second
	trieSyncer, _ := trie.NewTrieSyncer(trieNodesResolver, nRequester.ShardDataPool.TrieNodes(), requesterStateTrie, waitTime)
	err = trieSyncer.StartSyncing(rootHash)
	assert.Nil(t, err)

	newRootHash, _ := requesterStateTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)
}
