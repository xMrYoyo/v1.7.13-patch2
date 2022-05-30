package transactionAPI

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	coreMock "github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	dblookupextMock "github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgAPIBlockProcessor() *ArgAPITransactionProcessor {
	return &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &mock.PubkeyConverterMock{},
		ShardCoordinator:         createShardCoordinator(),
		HistoryRepository:        &dblookupextMock.HistoryRepositoryStub{},
		StorageService:           &mock.ChainStorerMock{},
		DataPool:                 &dataRetrieverMock.PoolsHolderMock{},
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              &testscommon.FeeComputerStub{},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
	}
}

func TestNewAPITransactionProcessor(t *testing.T) {
	t.Parallel()

	t.Run("NilArg", func(t *testing.T) {
		t.Parallel()

		_, err := NewAPITransactionProcessor(nil)
		require.Equal(t, ErrNilAPITransactionProcessorArg, err)
	})

	t.Run("NilMarshalizer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.Marshalizer = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("NilDataPool", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.DataPool = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilDataPoolHolder, err)
	})

	t.Run("NilHistoryRepository", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.HistoryRepository = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilHistoryRepository, err)
	})

	t.Run("NilShardCoordinator", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.ShardCoordinator = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.AddressPubKeyConverter = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilPubkeyConverter, err)
	})

	t.Run("NilStorageService", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.StorageService = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilStorage, err)
	})

	t.Run("NilUint64Converter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.Uint64ByteSliceConverter = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilUint64Converter, err)
	})

	t.Run("NilTxFeeComputer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.FeeComputer = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, ErrNilFeeComputer, err)
	})

	t.Run("NilTypeHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.TxTypeHandler = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilTxTypeHandler, err)
	})

	t.Run("NilLogsFacade", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPIBlockProcessor()
		arguments.LogsFacade = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, ErrNilLogsFacade, err)
	})
}

func TestNode_GetTransactionInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _, _, _ := createAPITransactionProc(t, 0, false)
	_, err := n.GetTransaction("zzz", false)
	assert.Error(t, err)
}

func TestNode_GetTransactionFromPool(t *testing.T) {
	t.Parallel()

	n, _, dataPool, _ := createAPITransactionProc(t, 42, false)

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	dataPool.Transactions().AddData([]byte("a"), txA, 42, "1")
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	dataPool.Transactions().AddData([]byte("b"), txB, 42, "1")
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	dataPool.Transactions().AddData([]byte("c"), txC, 42, "1")

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, uint32(1), actualA.SourceShard)
	require.Equal(t, uint32(2), actualA.DestinationShard)

	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, uint32(2), actualB.SourceShard)
	require.Equal(t, uint32(1), actualB.DestinationShard)

	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, uint32(1), actualC.SourceShard)
	require.Equal(t, uint32(1), actualC.DestinationShard)

	require.Equal(t, transaction.TxStatusPending, actualA.Status)
	require.Equal(t, transaction.TxStatusPending, actualB.Status)
	require.Equal(t, transaction.TxStatusPending, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	dataPool.RewardTransactions().AddData([]byte("d"), txD, 42, "foo")

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusPending, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	dataPool.UnsignedTransactions().AddData([]byte("e"), txE, 42, "foo")
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	dataPool.UnsignedTransactions().AddData([]byte("f"), txF, 42, "foo")
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	dataPool.UnsignedTransactions().AddData([]byte("g"), txG, 42, "foo")

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)
	require.Equal(t, transaction.TxStatusPending, actualF.Status)
	require.Equal(t, transaction.TxStatusPending, actualG.Status)
}

func TestNode_GetTransactionFromStorage(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, _ := createAPITransactionProc(t, 0, false)

	// Cross-shard, we are source
	internalMarshalizer := &mock.MarshalizerFake{}
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, internalMarshalizer)
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, internalMarshalizer)
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, internalMarshalizer)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, transaction.TxStatusPending, actualA.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualB.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, internalMarshalizer)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusSuccess, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, internalMarshalizer)
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, internalMarshalizer)
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, internalMarshalizer)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualF.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualG.Status)

	// Missing transaction
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("missing")), false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction not found")
	require.Nil(t, tx)

	// Badly serialized transaction
	_ = chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")), false)
	require.NotNil(t, err)
	require.Nil(t, tx)
}

func TestNode_GetTransactionWithResultsFromStorage(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerFake{}
	txHash := hex.EncodeToString([]byte("txHash"))
	tx := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	scResultHash := []byte("scHash")
	scResult := &smartContractResult.SmartContractResult{
		OriginalTxHash: []byte("txHash"),
	}

	resultHashesByTxHash := &dblookupext.ResultsHashesByTxHash{
		ScResultsHashesAndEpoch: []*dblookupext.ScResultsHashesAndEpoch{
			{
				Epoch:           0,
				ScResultsHashes: [][]byte{scResultHash},
			},
		},
	}

	chainStorer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.TransactionUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(tx)
					},
				}
			case dataRetriever.UnsignedTransactionUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(scResult)
					},
				}
			case dataRetriever.TxLogsUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return nil, errors.New("dummy")
					},
				}
			default:
				return nil
			}
		},
	}

	historyRepo := &dblookupextMock.HistoryRepositoryStub{
		GetMiniblockMetadataByTxHashCalled: func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
			return &dblookupext.MiniblockMetadata{}, nil
		},
		GetEventsHashesByTxHashCalled: func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return resultHashesByTxHash, nil
		},
	}

	feeComputer := &testscommon.FeeComputerStub{
		ComputeTransactionFeeCalled: func(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
			return big.NewInt(1000)
		},
	}

	args := &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &mock.PubkeyConverterMock{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		HistoryRepository:        historyRepo,
		StorageService:           chainStorer,
		DataPool:                 dataRetrieverMock.NewPoolsHolderMock(),
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              feeComputer,
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
	}
	apiTransactionProc, _ := NewAPITransactionProcessor(args)

	expectedTx := &transaction.ApiTransactionResult{
		Tx:                          &transaction.Transaction{Nonce: tx.Nonce, RcvAddr: tx.RcvAddr, SndAddr: tx.SndAddr, Value: tx.Value},
		Hash:                        "747848617368",
		ProcessingTypeOnSource:      process.MoveBalance.String(),
		ProcessingTypeOnDestination: process.MoveBalance.String(),
		Nonce:                       tx.Nonce,
		Receiver:                    hex.EncodeToString(tx.RcvAddr),
		Sender:                      hex.EncodeToString(tx.SndAddr),
		Status:                      transaction.TxStatusSuccess,
		MiniBlockType:               block.TxBlock.String(),
		Type:                        "normal",
		Value:                       "<nil>",
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				Hash:           hex.EncodeToString(scResultHash),
				OriginalTxHash: txHash,
			},
		},
		InitiallyPaidFee: "1000",
	}

	apiTx, err := apiTransactionProc.GetTransaction(txHash, true)
	require.Nil(t, err)
	require.Equal(t, expectedTx, apiTx)
}

func TestNode_lookupHistoricalTransaction(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, historyRepo := createAPITransactionProc(t, 42, true)

	// Normal transactions

	// Cross-shard, we are source
	internalMarshalizer := n.marshalizer
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 2, 42, nil, 0)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, 42, int(actualA.Epoch))
	require.Equal(t, transaction.TxStatusPending, actualA.Status)

	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 2, 1, 42, nil, 0)

	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, 42, int(actualB.Epoch))
	require.Equal(t, transaction.TxStatusSuccess, actualB.Status)

	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 1, 42, nil, 0)

	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, 42, int(actualC.Epoch))
	require.Equal(t, transaction.TxStatusSuccess, actualC.Status)

	// Invalid transaction
	txInvalid := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("invalid"), txInvalid, n.marshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.InvalidBlock, 1, 1, 42, nil, 0)

	actualInvalid, err := n.GetTransaction(hex.EncodeToString([]byte("invalid")), false)
	require.Nil(t, err)
	require.Equal(t, txInvalid.Nonce, actualInvalid.Nonce)
	require.Equal(t, 42, int(actualInvalid.Epoch))
	require.Equal(t, string(transaction.TxTypeInvalid), actualInvalid.Type)
	require.Equal(t, transaction.TxStatusInvalid, actualInvalid.Status)

	// Reward transactions
	headerHash := []byte("hash")
	headerNonce := uint64(1)
	nonceBytes := n.uint64ByteSliceConverter.ToByteSlice(headerNonce)
	_ = chainStorer.MetaHdrNonce.Put(nonceBytes, headerHash)
	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.RewardsBlock, core.MetachainShardId, 1, 42, headerHash, headerNonce)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualD.Epoch))
	require.Equal(t, string(transaction.TxTypeReward), actualD.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 2, 42, nil, 0)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualE.Epoch))
	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualE.Type)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)

	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 2, 1, 42, nil, 0)

	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualF.Epoch))
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualF.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualF.Status)

	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 1, 42, nil, 0)

	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualG.Epoch))
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualG.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualG.Status)

	// Missing transaction
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return nil, fmt.Errorf("fooError")
	}
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, tx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction not found")
	require.Contains(t, err.Error(), "fooError")

	// Badly serialized transaction
	_ = chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{}, nil
	}
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")), false)
	require.NotNil(t, err)
	require.Nil(t, tx)

	// Reward reverted transaction
	wrongHeaderHash := []byte("wrong-hash")
	headerHash = []byte("hash")
	headerNonce = uint64(1)
	nonceBytes = n.uint64ByteSliceConverter.ToByteSlice(headerNonce)
	_ = chainStorer.MetaHdrNonce.Put(nonceBytes, headerHash)
	txH := &rewardTx.RewardTx{Round: 50, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("h"), txH, n.marshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.RewardsBlock, core.MetachainShardId, 1, 42, wrongHeaderHash, headerNonce)

	actualH, err := n.GetTransaction(hex.EncodeToString([]byte("h")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualD.Epoch))
	require.Equal(t, string(transaction.TxTypeReward), actualH.Type)
	require.Equal(t, transaction.TxStatusRewardReverted, actualH.Status)
}

func TestNode_PutHistoryFieldsInTransaction(t *testing.T) {
	tx := &transaction.ApiTransactionResult{}
	metadata := &dblookupext.MiniblockMetadata{
		Epoch:                             42,
		Round:                             4321,
		MiniblockHash:                     []byte{15},
		DestinationShardID:                12,
		SourceShardID:                     11,
		HeaderNonce:                       4300,
		HeaderHash:                        []byte{14},
		NotarizedAtSourceInMetaNonce:      4250,
		NotarizedAtSourceInMetaHash:       []byte{13},
		NotarizedAtDestinationInMetaNonce: 4253,
		NotarizedAtDestinationInMetaHash:  []byte{12},
	}

	putMiniblockFieldsInTransaction(tx, metadata)

	require.Equal(t, 42, int(tx.Epoch))
	require.Equal(t, 4321, int(tx.Round))
	require.Equal(t, "0f", tx.MiniBlockHash)
	require.Equal(t, 12, int(tx.DestinationShard))
	require.Equal(t, 11, int(tx.SourceShard))
	require.Equal(t, 4300, int(tx.BlockNonce))
	require.Equal(t, "0e", tx.BlockHash)
	require.Equal(t, 4250, int(tx.NotarizedAtSourceInMetaNonce))
	require.Equal(t, "0d", tx.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4253, int(tx.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, "0c", tx.NotarizedAtDestinationInMetaHash)
}

func TestApiTransactionProcessor_GetTransactionsPool(t *testing.T) {
	t.Parallel()

	txHash0, txHash1, txHash2, txHash3 := []byte("txHash0"), []byte("txHash1"), []byte("txHash2"), []byte("txHash3")
	expectedTxs := [][]byte{txHash0, txHash1}
	expectedScrs := [][]byte{txHash2}
	expectedRwds := [][]byte{txHash3}
	args := createMockArgAPIBlockProcessor()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedTxs
				},
			}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedScrs
				},
			}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedRwds
				},
			}
		},
	}
	atp, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, atp)

	res, err := atp.GetTransactionsPool()
	require.NoError(t, err)
	require.Equal(t, []string{hex.EncodeToString(txHash0), hex.EncodeToString(txHash1)}, res.RegularTransactions)
	require.Equal(t, []string{hex.EncodeToString(txHash2)}, res.SmartContractResults)
	require.Equal(t, []string{hex.EncodeToString(txHash3)}, res.Rewards)
}

func createAPITransactionProc(t *testing.T, epoch uint32, withDbLookupExt bool) (*apiTransactionProcessor, *genericMocks.ChainStorerMock, *dataRetrieverMock.PoolsHolderMock, *dblookupextMock.HistoryRepositoryStub) {
	chainStorer := genericMocks.NewChainStorerMock(epoch)
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	historyRepo := &dblookupextMock.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return withDbLookupExt
		},
	}

	args := &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &mock.PubkeyConverterMock{},
		ShardCoordinator:         createShardCoordinator(),
		HistoryRepository:        historyRepo,
		StorageService:           chainStorer,
		DataPool:                 dataPool,
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              &testscommon.FeeComputerStub{},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
	}
	apiTransactionProc, err := NewAPITransactionProcessor(args)
	require.Nil(t, err)

	return apiTransactionProc, chainStorer, dataPool, historyRepo
}

func createShardCoordinator() *mock.ShardCoordinatorMock {
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfShardId: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			if address == nil {
				return core.MetachainShardId
			}
			if bytes.Equal(address, []byte("alice")) {
				return 1
			}
			if bytes.Equal(address, []byte("bob")) {
				return 2
			}
			panic("bad test")
		},
	}

	return shardCoordinator
}

func setupGetMiniblockMetadataByTxHash(
	historyRepo *dblookupextMock.HistoryRepositoryStub,
	blockType block.Type,
	sourceShard uint32,
	destinationShard uint32,
	epoch uint32,
	headerHash []byte,
	headerNonce uint64,
) {
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{
			Type:               int32(blockType),
			SourceShardID:      sourceShard,
			DestinationShardID: destinationShard,
			Epoch:              epoch,
			HeaderNonce:        headerNonce,
			HeaderHash:         headerHash,
		}, nil
	}
}

func TestPrepareUnsignedTx(t *testing.T) {
	t.Parallel()
	addrSize := 32
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          1,
		Value:          big.NewInt(2),
		SndAddr:        bytes.Repeat([]byte{0}, addrSize),
		RcvAddr:        bytes.Repeat([]byte{1}, addrSize),
		OriginalSender: []byte("invalid original sender"),
	}

	n, _, _, _ := createAPITransactionProc(t, 0, true)
	n.txUnmarshaller.addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addrSize, &coreMock.LoggerMock{})
	n.addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addrSize, &coreMock.LoggerMock{})

	scrResult1, err := n.txUnmarshaller.prepareUnsignedTx(scr1)
	assert.Nil(t, err)
	expectedScr1 := &transaction.ApiTransactionResult{
		Tx:             scr1,
		Nonce:          1,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "2",
		Receiver:       "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
		Sender:         "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
		OriginalSender: "",
	}
	assert.Equal(t, scrResult1, expectedScr1)

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          3,
		Value:          big.NewInt(4),
		SndAddr:        bytes.Repeat([]byte{5}, addrSize),
		RcvAddr:        bytes.Repeat([]byte{6}, addrSize),
		OriginalSender: bytes.Repeat([]byte{7}, addrSize),
	}

	scrResult2, err := n.txUnmarshaller.prepareUnsignedTx(scr2)
	assert.Nil(t, err)
	expectedScr2 := &transaction.ApiTransactionResult{
		Tx:             scr2,
		Nonce:          3,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "4",
		Receiver:       "erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e",
		Sender:         "erd1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zsrqsks3",
		OriginalSender: "erd1qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurstywtnm",
	}
	assert.Equal(t, scrResult2, expectedScr2)
}

func TestNode_ComputeTimestampForRound(t *testing.T) {
	genesis := getTime(t, "1596117600")
	n, _, _, _ := createAPITransactionProc(t, 0, false)
	n.genesisTime = genesis
	n.roundDuration = 6000

	res := n.computeTimestampForRound(0)
	require.Equal(t, int64(0), res)

	res = n.computeTimestampForRound(4837403)
	require.Equal(t, int64(1625142018), res)
}

func getTime(t *testing.T, timestamp string) time.Time {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		require.NoError(t, err)
	}
	tm := time.Unix(i, 0)

	return tm
}

func TestApiTransactionProcessor_GetTransactionPopulatesComputedFields(t *testing.T) {
	dataPool := dataRetrieverMock.NewPoolsHolderMock()
	feeComputer := &testscommon.FeeComputerStub{}
	txTypeHandler := &testscommon.TxTypeHandlerMock{}

	arguments := createMockArgAPIBlockProcessor()
	arguments.DataPool = dataPool
	arguments.FeeComputer = feeComputer
	arguments.TxTypeHandler = txTypeHandler

	processor, err := NewAPITransactionProcessor(arguments)
	require.Nil(t, err)
	require.NotNil(t, processor)

	t.Run("InitiallyPaidFee", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
			return big.NewInt(1000)
		}

		dataPool.Transactions().AddData([]byte{0, 0}, &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}, 42, "1")
		tx, err := processor.GetTransaction("0000", true)

		require.Nil(t, err)
		require.Equal(t, "1000", tx.InitiallyPaidFee)
	})

	t.Run("InitiallyPaidFee (missing on unsigned transaction)", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
			return big.NewInt(1000)
		}

		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@ok"), Value: big.NewInt(0)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 1}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0001", true)

		require.Nil(t, err)
		require.Equal(t, "", tx.InitiallyPaidFee)
	})

	t.Run("ProcessingType", func(t *testing.T) {
		txTypeHandler.ComputeTransactionTypeCalled = func(data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCDeployment
		}

		dataPool.Transactions().AddData([]byte{0, 2}, &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}, 42, "1")
		tx, err := processor.GetTransaction("0002", true)

		require.Nil(t, err)
		require.Equal(t, process.MoveBalance.String(), tx.ProcessingTypeOnSource)
		require.Equal(t, process.SCDeployment.String(), tx.ProcessingTypeOnDestination)
	})

	t.Run("IsRefund (false)", func(t *testing.T) {
		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@ok"), Value: big.NewInt(0)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 3}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0003", true)

		require.Nil(t, err)
		require.Equal(t, false, tx.IsRefund)
	})

	t.Run("IsRefund (true)", func(t *testing.T) {
		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@6f6b"), Value: big.NewInt(500)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 4}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0004", true)

		require.Nil(t, err)
		require.Equal(t, true, tx.IsRefund)
	})
}

func TestApiTransactionProcessor_UnmarshalTransactionPopulatesComputedFields(t *testing.T) {
	feeComputer := &testscommon.FeeComputerStub{}
	txTypeHandler := &testscommon.TxTypeHandlerMock{}

	arguments := createMockArgAPIBlockProcessor()
	arguments.Marshalizer = &marshal.GogoProtoMarshalizer{}
	arguments.FeeComputer = feeComputer
	arguments.TxTypeHandler = txTypeHandler

	processor, err := NewAPITransactionProcessor(arguments)
	require.Nil(t, err)
	require.NotNil(t, processor)

	t.Run("InitiallyPaidFee", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
			return big.NewInt(1000)
		}

		txBytes, err := hex.DecodeString("08061209000de0b6b3a76400001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d06520d6c6f63616c2d746573746e657458016240e011a7ab7788e40e61348445e2ccb55b0c61ab81d2ba88fda9d2d23b0a7512a627e2dc9b88bebcfdc4c49e9eaa2f65c016bc62ec3155dc3f60628cc7260e150d")
		require.Nil(t, err)

		tx, err := processor.UnmarshalTransaction(txBytes, transaction.TxTypeNormal)
		require.Nil(t, err)
		require.Equal(t, "1000", tx.InitiallyPaidFee)
	})

	t.Run("InitiallyPaidFee (missing on unsigned transaction)", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
			return big.NewInt(1000)
		}

		txBytes, err := hex.DecodeString("080712070021eca426ba801a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e12220000000000000000005004888d06daef6d4ce8a01d72812d08617b4b504a369e1320100420540366636624a205be93498d366ab14a6794c5c5661c06e70cfef2fbfbd460911c6c924703594ef52205be93498d366ab14a6794c5c5661c06e70cfef2fbfbd460911c6c924703594ef608094ebdc03")
		require.Nil(t, err)

		tx, err := processor.UnmarshalTransaction(txBytes, transaction.TxTypeUnsigned)
		require.Nil(t, err)
		require.Equal(t, "", tx.InitiallyPaidFee)
	})

	t.Run("ProcessingType", func(t *testing.T) {
		txTypeHandler.ComputeTransactionTypeCalled = func(data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCDeployment
		}

		txBytes, err := hex.DecodeString("08061209000de0b6b3a76400001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d06520d6c6f63616c2d746573746e657458016240e011a7ab7788e40e61348445e2ccb55b0c61ab81d2ba88fda9d2d23b0a7512a627e2dc9b88bebcfdc4c49e9eaa2f65c016bc62ec3155dc3f60628cc7260e150d")
		require.Nil(t, err)

		tx, err := processor.UnmarshalTransaction(txBytes, transaction.TxTypeNormal)
		require.Nil(t, err)
		require.Equal(t, process.MoveBalance.String(), tx.ProcessingTypeOnSource)
		require.Equal(t, process.SCDeployment.String(), tx.ProcessingTypeOnDestination)
	})

	t.Run("IsRefund (false)", func(t *testing.T) {
		txBytes, err := hex.DecodeString("08061209000de0b6b3a76400001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d06520d6c6f63616c2d746573746e657458016240e011a7ab7788e40e61348445e2ccb55b0c61ab81d2ba88fda9d2d23b0a7512a627e2dc9b88bebcfdc4c49e9eaa2f65c016bc62ec3155dc3f60628cc7260e150d")
		require.Nil(t, err)

		tx, err := processor.UnmarshalTransaction(txBytes, transaction.TxTypeNormal)
		require.Nil(t, err)
		require.Equal(t, false, tx.IsRefund)
	})

	t.Run("IsRefund (true)", func(t *testing.T) {
		txBytes, err := hex.DecodeString("080712070021eca426ba801a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e12220000000000000000005004888d06daef6d4ce8a01d72812d08617b4b504a369e1320100420540366636624a205be93498d366ab14a6794c5c5661c06e70cfef2fbfbd460911c6c924703594ef52205be93498d366ab14a6794c5c5661c06e70cfef2fbfbd460911c6c924703594ef608094ebdc03")
		require.Nil(t, err)

		tx, err := processor.UnmarshalTransaction(txBytes, transaction.TxTypeUnsigned)
		require.Nil(t, err)
		require.Equal(t, true, tx.IsRefund)
	})
}
