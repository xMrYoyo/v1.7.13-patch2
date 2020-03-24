package sharding

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	keyFormat = "%s_%v_%v_%v"
)

// TODO: move this to config parameters
const nodeCoordinatorStoredEpochs = 2

type validatorWithShardID struct {
	validator Validator
	shardID   uint32
}

// TODO: add a parameter for shardID  when acting as observer
type epochNodesConfig struct {
	nbShards                uint32
	shardID                 uint32
	eligibleMap             map[uint32][]Validator
	waitingMap              map[uint32][]Validator
	expandedEligibleMap     map[uint32][]Validator
	publicKeyToValidatorMap map[string]*validatorWithShardID
	mutNodesMaps            sync.RWMutex
}

// EpochStartEventNotifier provides Register and Unregister functionality for the end of epoch events
type EpochStartEventNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
}

type indexHashedNodesCoordinator struct {
	hasher                        hashing.Hasher
	shuffler                      NodesShuffler
	epochStartRegistrationHandler EpochStartEventNotifier
	listIndexUpdater              ListIndexUpdaterHandler
	bootStorer                    storage.Storer
	selfPubKey                    []byte
	nodesConfig                   map[uint32]*epochNodesConfig
	mutNodesConfig                sync.RWMutex
	currentEpoch                  uint32
	savedStateKey                 []byte
	mutSavedStateKey              sync.RWMutex
	numTotalEligible              uint64
	shardConsensusGroupSize       int
	metaConsensusGroupSize        int
	nodesPerShardSetter           NodesPerShardSetter
	consensusGroupCacher          Cacher
	shardIDAsObserver       uint32
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*epochNodesConfig, nodeCoordinatorStoredEpochs)

	nodesConfig[arguments.Epoch] = &epochNodesConfig{
		nbShards:     arguments.NbShards,
		shardID:      arguments.ShardIDAsObserver,
		eligibleMap:  make(map[uint32][]Validator),
		waitingMap:   make(map[uint32][]Validator),
		mutNodesMaps: sync.RWMutex{},
	}

	savedKey := arguments.Hasher.Compute(string(arguments.SelfPublicKey))

	ihgs := &indexHashedNodesCoordinator{
		hasher:                        arguments.Hasher,
		shuffler:                      arguments.Shuffler,
		epochStartRegistrationHandler: arguments.EpochStartNotifier,
		bootStorer:                    arguments.BootStorer,
		listIndexUpdater:              arguments.ListIndexUpdater,
		selfPubKey:                    arguments.SelfPublicKey,
		nodesConfig:                   nodesConfig,
		currentEpoch:                  arguments.Epoch,
		savedStateKey:                 savedKey,
		shardConsensusGroupSize:       arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:        arguments.MetaConsensusGroupSize,
		consensusGroupCacher:          arguments.ConsensusGroupCache,
		shardIDAsObserver:       	   arguments.ShardIDAsObserver,
	}

	ihgs.nodesPerShardSetter = ihgs
	err = ihgs.nodesPerShardSetter.SetNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, arguments.Epoch, false)
	if err != nil {
		return nil, err
	}

	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

	ihgs.epochStartRegistrationHandler.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if arguments.ShardConsensusGroupSize < 1 || arguments.MetaConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if arguments.NbShards < 1 {
		return ErrInvalidNumberOfShards
	}
	if arguments.ShardIDAsObserver >= arguments.NbShards && arguments.ShardIDAsObserver != core.MetachainShardId {
		return ErrInvalidShardId
	}
	if check.IfNil(arguments.Hasher) {
		return ErrNilHasher
	}
	if len(arguments.SelfPublicKey) == 0 {
		return ErrNilPubKey
	}
	if check.IfNil(arguments.Shuffler) {
		return ErrNilShuffler
	}
	if check.IfNil(arguments.ListIndexUpdater) {
		return ErrNilListIndexUpdater
	}
	if check.IfNil(arguments.BootStorer) {
		return ErrNilBootStorer
	}
	if arguments.ConsensusGroupCache == nil {
		return ErrNilCacher
	}

	return nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinator) SetNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	epoch uint32,
	updatePeersListAndIndex bool,
) error {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nodesList, ok := eligible[core.MetachainShardId]
	if !ok || len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	numTotalEligible := uint64(len(nodesList))
	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
		numTotalEligible += uint64(nbNodesShard)
	}

	// nbShards holds number of shards without meta
	nodesConfig.nbShards = uint32(len(eligible) - 1)
	nodesConfig.eligibleMap = eligible
	nodesConfig.waitingMap = waiting
	nodesConfig.publicKeyToValidatorMap = make(map[string]*validatorWithShardID)
	for shardId, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			nodesConfig.publicKeyToValidatorMap[string(shardEligible[i].PubKey())] = &validatorWithShardID{
				validator: shardEligible[i],
				shardID:   shardId,
			}
		}
	}
	for shardId, shardWaiting := range nodesConfig.waitingMap {
		for i := 0; i < len(shardWaiting); i++ {
			nodesConfig.publicKeyToValidatorMap[string(shardWaiting[i].PubKey())] = &validatorWithShardID{
				validator: shardWaiting[i],
				shardID:   shardId,
			}
		}
	}

	nodesConfig.shardID = ihgs.computeShardForSelfPublicKey(nodesConfig)
	ihgs.nodesConfig[epoch] = nodesConfig
	ihgs.numTotalEligible = numTotalEligible

	if updatePeersListAndIndex {
		err := ihgs.updatePeersListAndIndex(nodesConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

// ComputeLeaving -
func (ihgs *indexHashedNodesCoordinator) ComputeLeaving([]Validator) []Validator {
	return make([]Validator, 0)
}

// ComputeConsensusGroup will generate a list of validators based on the the eligible list,
// consensus group size and a randomness source
// Steps:
// 1. generate expanded eligible list by multiplying entries from shards' eligible list according to stake and rating -> TODO
// 2. for each value in [0, consensusGroupSize), compute proposedindex = Hash( [index as string] CONCAT randomness) % len(eligible list)
// 3. if proposed index is already in the temp validator list, then proposedIndex++ (and then % len(eligible list) as to not
//    exceed the maximum index value permitted by the validator list), and then recheck against temp validator list until
//    the item at the new proposed index is not found in the list. This new proposed index will be called checked index
// 4. the item at the checked index is appended in the temp validator list
func (ihgs *indexHashedNodesCoordinator) ComputeConsensusGroup(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) (validatorsGroup []Validator, err error) {
	var expandedList []Validator

	log.Trace("computing consensus group for",
		"epoch", epoch,
		"shardID", shardID,
		"randomness", randomness,
		"round", round)

	if len(randomness) == 0 {
		return nil, ErrNilRandomness
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if ok {
		if shardID >= nodesConfig.nbShards && shardID != core.MetachainShardId {
			return nil, ErrInvalidShardId
		}
		expandedList = nodesConfig.eligibleMap[shardID]
	}
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardID, epoch))
	validators := ihgs.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihgs.ConsensusGroupSize(shardID)
	randomness = []byte(fmt.Sprintf("%d-%s", round, randomness))

	log.Debug("ComputeValidatorsGroup",
		"randomness", randomness,
		"consensus size", consensusSize,
		"eligible list length", len(expandedList))

	consensusGroupProvider := NewSelectionBasedProvider(ihgs.hasher, uint32(consensusSize))
	tempList, err := consensusGroupProvider.Get(randomness, int64(consensusSize), expandedList)
	if err != nil {
		return nil, err
	}

	ihgs.consensusGroupCacher.Put(key, tempList)

	return tempList, nil
}

func (ihgs *indexHashedNodesCoordinator) searchConsensusForKey(key []byte) []Validator {
	value, ok := ihgs.consensusGroupCacher.Get(key)
	if ok {
		consensusGroup, typeOk := value.([]Validator)
		if typeOk {
			return consensusGroup
		}

	}
	return nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihgs *indexHashedNodesCoordinator) GetValidatorWithPublicKey(
	publicKey []byte,
	epoch uint32,
) (Validator, uint32, error) {
	if len(publicKey) == 0 {
		return nil, 0, ErrNilPubKey
	}
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, 0, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	v, ok := nodesConfig.publicKeyToValidatorMap[string(publicKey)]
	if ok {
		return v.validator, v.shardID, nil
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *indexHashedNodesCoordinator) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardID uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeConsensusGroup(randomness, round, shardID, epoch)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetAllEligibleValidatorsPublicKeys will return all validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], nodesConfig.eligibleMap[shardID][i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetAllWaitingValidatorsPublicKeys will return all validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardID, shardEligible := range nodesConfig.waitingMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardID] = append(validatorsPubKeys[shardID], nodesConfig.waitingMap[shardID][i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetWaitingPublicKeysPerShard will return all validators public keys in waiting list for all shards
func (ihgs *indexHashedNodesCoordinator) GetWaitingPublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardId, shardWaiting := range nodesConfig.waitingMap {
		for i := 0; i < len(shardWaiting); i++ {
			validatorsPubKeys[shardId] = append(validatorsPubKeys[shardId], nodesConfig.waitingMap[shardId][i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetValidatorsIndexes will return validators indexes for a block
func (ihgs *indexHashedNodesCoordinator) GetValidatorsIndexes(
	publicKeys []string,
	epoch uint32,
) ([]uint64, error) {
	signersIndexes := make([]uint64, 0)

	validatorsPubKeys, err := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	for _, pubKey := range publicKeys {
		for index, value := range validatorsPubKeys[nodesConfig.shardID] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	if len(publicKeys) != len(signersIndexes) {
		strHaving := "having the following keys: \n"
		for index, value := range validatorsPubKeys[nodesConfig.shardID] {
			strHaving += fmt.Sprintf(" index %d  key %s\n", index, display.DisplayByteSlice(value))
		}

		strNeeded := "needed the following keys: \n"
		for _, pubKey := range publicKeys {
			strNeeded += fmt.Sprintf(" key %s\n", display.DisplayByteSlice([]byte(pubKey)))
		}

		log.Error("public keys not found\n"+strHaving+"\n"+strNeeded+"\n",
			"len pubKeys", len(publicKeys),
			"len signers", len(signersIndexes),
		)

		return nil, ErrInvalidNumberPubKeys
	}

	return signersIndexes, nil
}

// EpochStartPrepare is called when an epoch start event is observed, but not yet confirmed/committed.
// Some components may need to do some initialisation on this event
func (ihgs *indexHashedNodesCoordinator) EpochStartPrepare(metaHeader data.HeaderHandler) {
	randomness := metaHeader.GetPrevRandSeed()
	newEpoch := metaHeader.GetEpoch()

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[newEpoch-1]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		log.Error("no configured epoch found")
		return
	}

	allValidators := make([]Validator, 0)

	for _, shardValidators := range nodesConfig.eligibleMap {
		allValidators = append(allValidators, shardValidators...)
	}

	for _, shardValidators := range nodesConfig.waitingMap {
		allValidators = append(allValidators, shardValidators...)
	}

	sort.Slice(allValidators, func(i, j int) bool {
		return bytes.Compare(allValidators[i].PubKey(), allValidators[j].PubKey()) < 0
	})

	leaving := ihgs.nodesPerShardSetter.ComputeLeaving(allValidators)

	// TODO: update the new nodes and leaving nodes as well
	shufflerArgs := ArgsUpdateNodes{
		Eligible: nodesConfig.eligibleMap,
		Waiting:  nodesConfig.waitingMap,
		NewNodes: make([]Validator, 0),
		Leaving:  leaving,
		Rand:     randomness,
		NbShards: nodesConfig.nbShards,
	}

	eligibleMap, waitingMap, stillRemaining := ihgs.shuffler.UpdateNodeLists(shufflerArgs)

	err := ihgs.nodesPerShardSetter.SetNodesPerShards(eligibleMap, waitingMap, newEpoch, true)
	if err != nil {
		log.Error("set nodes per shard failed", "error", err.Error())
	}

	err = ihgs.saveState(randomness)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	displayNodesConfiguration(eligibleMap, waitingMap, leaving, stillRemaining)

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = randomness
	ihgs.mutSavedStateKey.Unlock()
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodeCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0
	ihgs.currentEpoch = newEpoch

	err := ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}

	ihgs.mutNodesConfig.Lock()
	if needToRemove {
		for epoch := range ihgs.nodesConfig {
			if epoch <= uint32(epochToRemove) {
				delete(ihgs.nodesConfig, epoch)
			}
		}
	}
	ihgs.mutNodesConfig.Unlock()
}

// NotifyOrder returns the notification order for a start of epoch event
func (ihgs *indexHashedNodesCoordinator) NotifyOrder() uint32 {
	return core.NodesCoordinatorOrder
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
func (ihgs *indexHashedNodesCoordinator) GetSavedStateKey() []byte {
	ihgs.mutSavedStateKey.RLock()
	key := ihgs.savedStateKey
	ihgs.mutSavedStateKey.RUnlock()

	return key
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihgs *indexHashedNodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return 0, ErrEpochNodesConfigDoesNotExist
	}

	return nodesConfig.shardID, nil
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ihgs *indexHashedNodesCoordinator) GetConsensusWhitelistedNodes(
	epoch uint32,
) (map[string]struct{}, error) {
	var err error
	shardEligible := make(map[string]struct{})
	publicKeysPrevEpoch := make(map[uint32][][]byte)
	prevEpochConfigExists := false

	if epoch > 0 {
		publicKeysPrevEpoch, err = ihgs.GetAllEligibleValidatorsPublicKeys(epoch - 1)
		if err == nil {
			prevEpochConfigExists = true
		} else {
			log.Warn("get consensus whitelisted nodes", "error", err.Error())
		}
	}

	var prevEpochShardId uint32
	if prevEpochConfigExists {
		prevEpochShardId, err = ihgs.ShardIdForEpoch(epoch - 1)
		if err == nil {
			for _, pubKey := range publicKeysPrevEpoch[prevEpochShardId] {
				shardEligible[string(pubKey)] = struct{}{}
			}
		} else {
			log.Trace("not critical error getting shardID for epoch", "epoch", epoch-1, "error", err)
		}
	}

	publicKeysNewEpoch, errGetEligible := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	if errGetEligible != nil {
		return nil, errGetEligible
	}

	epochShardId, errShardIdForEpoch := ihgs.ShardIdForEpoch(epoch)
	if errShardIdForEpoch != nil {
		return nil, errShardIdForEpoch
	}

	for _, pubKey := range publicKeysNewEpoch[epochShardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}

	return shardEligible, nil
}

// UpdatePeersListAndIndex will update the list and the index for all peers
func (ihgs *indexHashedNodesCoordinator) UpdatePeersListAndIndex() error {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return ErrEpochNodesConfigDoesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	return ihgs.updatePeersListAndIndex(nodesConfig)
}

// updatePeersListAndIndex will update the list and the index for all peers
// should be called with mutex locked
func (ihgs *indexHashedNodesCoordinator) updatePeersListAndIndex(nodesConfig *epochNodesConfig) error {
	err := ihgs.updatePeerAccountsForGivenMap(nodesConfig.eligibleMap, core.EligibleList)
	if err != nil {
		return err
	}

	err = ihgs.updatePeerAccountsForGivenMap(nodesConfig.waitingMap, core.WaitingList)
	if err != nil {
		return err
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinator) updatePeerAccountsForGivenMap(
	peers map[uint32][]Validator,
	list core.PeerType,
) error {
	for shardId, accountsPerShard := range peers {
		for index, account := range accountsPerShard {
			err := ihgs.listIndexUpdater.UpdateListAndIndex(
				string(account.PubKey()),
				shardId,
				string(list),
				int32(index))
			if err != nil {
				log.Warn("error while updating list and index for peer",
					"error", err,
					"public key", account.PubKey())
			}
		}
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinator) computeShardForSelfPublicKey(nodesConfig *epochNodesConfig) uint32 {
	pubKey := ihgs.selfPubKey
	selfShard := ihgs.shardIDAsObserver
	epNodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
	if ok {
		log.Trace("computeShardForSelfPublicKey found existing config",
			"shard", epNodesConfig.shardID,
		)
		selfShard = epNodesConfig.shardID
	}

	for shard, validators := range nodesConfig.eligibleMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				log.Trace("computeShardForSelfPublicKey found validator in eligible",
					"shard", shard,
					"validator PK", v,
				)

				return shard
			}
		}
	}

	for shard, validators := range nodesConfig.waitingMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				log.Trace("computeShardForSelfPublicKey found validator in waiting",
					"shard", shard,
					"validator PK", v,
				)

				return shard
			}
		}
	}

	log.Trace("computeShardForSelfPublicKey returned default",
		"shard", selfShard,
	)
	return selfShard
}

// ConsensusGroupSize returns the consensus group size for a specific shard
func (ihgs *indexHashedNodesCoordinator) ConsensusGroupSize(
	shardID uint32,
) int {
	if shardID == core.MetachainShardId {
		return ihgs.metaConsensusGroupSize
	}

	return ihgs.shardConsensusGroupSize
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihgs *indexHashedNodesCoordinator) GetNumTotalEligible() uint64 {
	return ihgs.numTotalEligible
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihgs *indexHashedNodesCoordinator) GetOwnPublicKey() []byte {
	return ihgs.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihgs == nil
}

func displayNodesConfiguration(eligible map[uint32][]Validator, waiting map[uint32][]Validator, leaving []Validator, actualLeaving []Validator) {
	for shardID, validators := range eligible {
		for _, v := range validators {
			pk := v.PubKey()
			log.Debug("eligible", "pk", pk, "shardID", shardID)
		}
	}

	for shardID, validators := range waiting {
		for _, v := range validators {
			pk := v.PubKey()
			log.Debug("waiting", "pk", pk, "shardID", shardID)
		}
	}

	for _, v := range leaving {
		pk := v.PubKey()
		log.Debug("computed leaving", "pk", pk)
	}

	for _, v := range actualLeaving {
		pk := v.PubKey()
		log.Debug("actually remaining", "pk", pk)
	}
}
