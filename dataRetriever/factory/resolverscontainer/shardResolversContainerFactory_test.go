package resolverscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected error")

func createStubTopicMessageHandlerForShard(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
	tmhs := mock.NewTopicMessageHandlerStub()

	tmhs.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}

		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	tmhs.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}

		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return tmhs
}

func createDataPoolsForShard() dataRetriever.PoolsHolder {
	pools := dataRetrieverMock.NewPoolsHolderStub()
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}

	return pools
}

func createStoreForShard() dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

func createTriesHolderForShard() common.TriesHolder {
	triesHolder := state.NewDataTriesHolder()
	triesHolder.Put([]byte(dataRetriever.UserAccountsUnit.String()), &trieMock.TrieStub{})
	triesHolder.Put([]byte(dataRetriever.PeerAccountsUnit.String()), &trieMock.TrieStub{})
	return triesHolder
}

// ------- NewResolversContainerFactory

func TestNewShardResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewShardResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Store = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshalizer = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilMarshalizerAndSizeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshalizer = nil
	args.SizeCheckDelta = 1
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardResolversContainerFactory_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.DataPools = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPoolHolder, err)
}

func TestNewShardResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Uint64ByteSliceConverter = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardResolversContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.DataPacker = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewShardResolversContainerFactory_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.PreferredPeersHolder = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewShardResolversContainerFactory_NilTriesContainerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.TriesContainer = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
}

func TestNewShardResolversContainerFactory_NilInputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.InputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

func TestNewShardResolversContainerFactory_NilOutputAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.OutputAntifloodHandler = nil
	rcf, err := resolverscontainer.NewShardResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.True(t, errors.Is(err, dataRetriever.ErrNilAntifloodHandler))
}

// ------- Create

func TestShardResolversContainerFactory_CreateRegisterTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.TransactionTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterHdrFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.ShardBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterMiniBlocksFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.MiniBlocksTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterTrieNodesFailsShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", factory.AccountTrieNodesTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateRegisterPeerAuthenticationShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = createStubTopicMessageHandlerForShard("", common.PeerAuthenticationTopic)
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.Nil(t, container)
	assert.Equal(t, errExpected, err)
}

func TestShardResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsShard()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	container, _ := rcf.Create()

	numResolverSCRs := noOfShards + 1
	numResolverTxs := noOfShards + 1
	numResolverRewardTxs := 1
	numResolverHeaders := 1
	numResolverMiniBlocks := noOfShards + 2
	numResolverMetaBlockHeaders := 1
	numResolverTrieNodes := 1
	numResolverPeerAuth := 1
	numResolverValidatorInfo := 1
	totalResolvers := numResolverTxs + numResolverHeaders + numResolverMiniBlocks + numResolverMetaBlockHeaders +
		numResolverSCRs + numResolverRewardTxs + numResolverTrieNodes + numResolverPeerAuth + numResolverValidatorInfo

	assert.Equal(t, totalResolvers, container.Len())
}

func getArgumentsShard() resolverscontainer.FactoryArgs {
	return resolverscontainer.FactoryArgs{
		ShardCoordinator:           mock.NewOneShardCoordinatorMock(),
		Messenger:                  createStubTopicMessageHandlerForShard("", ""),
		Store:                      createStoreForShard(),
		Marshalizer:                &mock.MarshalizerMock{},
		DataPools:                  createDataPoolsForShard(),
		Uint64ByteSliceConverter:   &mock.Uint64ByteSliceConverterMock{},
		DataPacker:                 &mock.DataPackerStub{},
		TriesContainer:             createTriesHolderForShard(),
		SizeCheckDelta:             0,
		InputAntifloodHandler:      &mock.P2PAntifloodHandlerStub{},
		OutputAntifloodHandler:     &mock.P2PAntifloodHandlerStub{},
		NumConcurrentResolvingJobs: 10,
		PreferredPeersHolder:       &p2pmocks.PeersHolderStub{},
		PayloadValidator:           &testscommon.PeerAuthenticationPayloadValidatorStub{},
	}
}
