package peer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

var _ process.ValidatorsProvider = (*validatorsProvider)(nil)

// validatorsProvider is the main interface for validators' provider
type validatorsProvider struct {
	nodesCoordinator             process.NodesCoordinator
	validatorStatistics          process.ValidatorStatisticsProcessor
	cache                        map[string]*state.ValidatorApiResponse
	cachedAuctionValidators      []*common.AuctionListValidatorAPIResponse
	cachedRandomness             []byte
	cacheRefreshIntervalDuration time.Duration
	refreshCache                 chan uint32
	lastCacheUpdate              time.Time
	lastAuctionCacheUpdate       time.Time
	lock                         sync.RWMutex
	auctionMutex                 sync.RWMutex
	cancelFunc                   func()
	validatorPubKeyConverter     core.PubkeyConverter
	addressPubKeyConverter       core.PubkeyConverter
	stakingDataProvider          epochStart.StakingDataProvider
	auctionListSelector          epochStart.AuctionListSelector

	maxRating    uint32
	currentEpoch uint32
}

// ArgValidatorsProvider contains all parameters needed for creating a validatorsProvider
type ArgValidatorsProvider struct {
	NodesCoordinator                  process.NodesCoordinator
	EpochStartEventNotifier           process.EpochStartEventNotifier
	CacheRefreshIntervalDurationInSec time.Duration
	ValidatorStatistics               process.ValidatorStatisticsProcessor
	ValidatorPubKeyConverter          core.PubkeyConverter
	AddressPubKeyConverter            core.PubkeyConverter
	StakingDataProvider               epochStart.StakingDataProvider
	AuctionListSelector               epochStart.AuctionListSelector
	StartEpoch                        uint32
	MaxRating                         uint32
}

// NewValidatorsProvider instantiates a new validatorsProvider structure responsible of keeping account of
//  the latest information about the validators
func NewValidatorsProvider(
	args ArgValidatorsProvider,
) (*validatorsProvider, error) {
	if check.IfNil(args.ValidatorStatistics) {
		return nil, process.ErrNilValidatorStatistics
	}
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for validators", process.ErrNilPubkeyConverter)
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, fmt.Errorf("%w for addresses", process.ErrNilPubkeyConverter)
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(args.EpochStartEventNotifier) {
		return nil, process.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.StakingDataProvider) {
		return nil, process.ErrNilStakingDataProvider
	}
	if check.IfNil(args.AuctionListSelector) {
		return nil, epochStart.ErrNilAuctionListSelector
	}
	if args.MaxRating == 0 {
		return nil, process.ErrMaxRatingZero
	}
	if args.CacheRefreshIntervalDurationInSec <= 0 {
		return nil, process.ErrInvalidCacheRefreshIntervalInSec
	}

	currentContext, cancelfunc := context.WithCancel(context.Background())

	valProvider := &validatorsProvider{
		nodesCoordinator:             args.NodesCoordinator,
		validatorStatistics:          args.ValidatorStatistics,
		stakingDataProvider:          args.StakingDataProvider,
		cache:                        make(map[string]*state.ValidatorApiResponse),
		cachedAuctionValidators:      make([]*common.AuctionListValidatorAPIResponse, 0),
		cachedRandomness:             make([]byte, 0),
		cacheRefreshIntervalDuration: args.CacheRefreshIntervalDurationInSec,
		refreshCache:                 make(chan uint32),
		lock:                         sync.RWMutex{},
		auctionMutex:                 sync.RWMutex{},
		cancelFunc:                   cancelfunc,
		maxRating:                    args.MaxRating,
		validatorPubKeyConverter:     args.ValidatorPubKeyConverter,
		addressPubKeyConverter:       args.AddressPubKeyConverter,
		currentEpoch:                 args.StartEpoch,
		auctionListSelector:          args.AuctionListSelector,
	}

	go valProvider.startRefreshProcess(currentContext)
	args.EpochStartEventNotifier.RegisterHandler(valProvider.epochStartEventHandler())

	return valProvider, nil
}

// GetLatestValidators gets the latest configuration of validators from the peerAccountsTrie
func (vp *validatorsProvider) GetLatestValidators() map[string]*state.ValidatorApiResponse {
	return vp.getValidators()
}

func (vp *validatorsProvider) getValidators() map[string]*state.ValidatorApiResponse {
	vp.lock.RLock()
	shouldUpdate := time.Since(vp.lastCacheUpdate) > vp.cacheRefreshIntervalDuration
	vp.lock.RUnlock()

	if shouldUpdate {
		vp.updateCache()
	}

	vp.lock.RLock()
	clonedMap := cloneMap(vp.cache)
	vp.lock.RUnlock()

	return clonedMap
}

func cloneMap(cache map[string]*state.ValidatorApiResponse) map[string]*state.ValidatorApiResponse {
	newMap := make(map[string]*state.ValidatorApiResponse)

	for k, v := range cache {
		newMap[k] = cloneValidatorAPIResponse(v)
	}

	return newMap
}

func cloneValidatorAPIResponse(v *state.ValidatorApiResponse) *state.ValidatorApiResponse {
	if v == nil {
		return nil
	}
	return &state.ValidatorApiResponse{
		TempRating:                         v.TempRating,
		NumLeaderSuccess:                   v.NumLeaderSuccess,
		NumLeaderFailure:                   v.NumLeaderFailure,
		NumValidatorSuccess:                v.NumValidatorSuccess,
		NumValidatorFailure:                v.NumValidatorFailure,
		NumValidatorIgnoredSignatures:      v.NumValidatorIgnoredSignatures,
		Rating:                             v.Rating,
		RatingModifier:                     v.RatingModifier,
		TotalNumLeaderSuccess:              v.TotalNumLeaderSuccess,
		TotalNumLeaderFailure:              v.TotalNumLeaderFailure,
		TotalNumValidatorSuccess:           v.TotalNumValidatorSuccess,
		TotalNumValidatorFailure:           v.TotalNumValidatorFailure,
		TotalNumValidatorIgnoredSignatures: v.TotalNumValidatorIgnoredSignatures,
		ShardId:                            v.ShardId,
		ValidatorStatus:                    v.ValidatorStatus,
	}
}

func (vp *validatorsProvider) epochStartEventHandler() nodesCoordinator.EpochStartActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		func(hdr data.HeaderHandler) {
			log.Trace("epochStartEventHandler - refreshCache forced",
				"nonce", hdr.GetNonce(),
				"shard", hdr.GetShardID(),
				"round", hdr.GetRound(),
				"epoch", hdr.GetEpoch())
			go func() {
				vp.refreshCache <- hdr.GetEpoch()
			}()
		},
		func(_ data.HeaderHandler) {},
		common.IndexerOrder,
	)

	return subscribeHandler
}

func (vp *validatorsProvider) startRefreshProcess(ctx context.Context) {
	for {
		vp.updateCache()
		err := vp.updateAuctionListCache()
		if err != nil {
			log.Error("could not update validators auction info cache", "error", err)
		}

		select {
		case epoch := <-vp.refreshCache:
			vp.lock.Lock()
			vp.currentEpoch = epoch
			vp.lock.Unlock()
			log.Trace("startRefreshProcess - forced refresh", "epoch", vp.currentEpoch)
		case <-ctx.Done():
			log.Debug("validatorsProvider's go routine is stopping...")
			return
		}
	}
}

func (vp *validatorsProvider) updateCache() {
	lastFinalizedRootHash := vp.validatorStatistics.LastFinalizedRootHash()
	if len(lastFinalizedRootHash) == 0 {
		return
	}
	allNodes, err := vp.validatorStatistics.GetValidatorInfoForRootHash(lastFinalizedRootHash)
	if err != nil {
		allNodes = state.NewShardValidatorsInfoMap()
		log.Trace("validatorsProvider - GetLatestValidatorInfos failed", "error", err)
	}

	vp.lock.RLock()
	epoch := vp.currentEpoch
	vp.lock.RUnlock()

	newCache := vp.createNewCache(epoch, allNodes)

	vp.lock.Lock()
	vp.lastCacheUpdate = time.Now()
	vp.cache = newCache
	vp.lock.Unlock()
}

func (vp *validatorsProvider) createNewCache(
	epoch uint32,
	allNodes state.ShardValidatorsInfoMapHandler,
) map[string]*state.ValidatorApiResponse {
	newCache := vp.createValidatorApiResponseMapFromValidatorInfoMap(allNodes)

	nodesMapEligible, err := vp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("validatorsProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregateLists(newCache, nodesMapEligible, common.EligibleList)

	nodesMapWaiting, err := vp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("validatorsProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregateLists(newCache, nodesMapWaiting, common.WaitingList)

	return newCache
}

func (vp *validatorsProvider) createValidatorApiResponseMapFromValidatorInfoMap(allNodes state.ShardValidatorsInfoMapHandler) map[string]*state.ValidatorApiResponse {
	newCache := make(map[string]*state.ValidatorApiResponse)

	for _, validatorInfo := range allNodes.GetAllValidatorsInfo() {
		strKey := vp.validatorPubKeyConverter.Encode(validatorInfo.GetPublicKey())
		newCache[strKey] = &state.ValidatorApiResponse{
			NumLeaderSuccess:                   validatorInfo.GetLeaderSuccess(),
			NumLeaderFailure:                   validatorInfo.GetLeaderFailure(),
			NumValidatorSuccess:                validatorInfo.GetValidatorSuccess(),
			NumValidatorFailure:                validatorInfo.GetValidatorFailure(),
			NumValidatorIgnoredSignatures:      validatorInfo.GetValidatorIgnoredSignatures(),
			TotalNumLeaderSuccess:              validatorInfo.GetTotalLeaderSuccess(),
			TotalNumLeaderFailure:              validatorInfo.GetTotalLeaderFailure(),
			TotalNumValidatorSuccess:           validatorInfo.GetTotalValidatorSuccess(),
			TotalNumValidatorFailure:           validatorInfo.GetTotalValidatorFailure(),
			TotalNumValidatorIgnoredSignatures: validatorInfo.GetTotalValidatorIgnoredSignatures(),
			RatingModifier:                     validatorInfo.GetRatingModifier(),
			Rating:                             float32(validatorInfo.GetRating()) * 100 / float32(vp.maxRating),
			TempRating:                         float32(validatorInfo.GetTempRating()) * 100 / float32(vp.maxRating),
			ShardId:                            validatorInfo.GetShardId(),
			ValidatorStatus:                    validatorInfo.GetList(),
		}
	}

	return newCache
}

func (vp *validatorsProvider) aggregateLists(
	newCache map[string]*state.ValidatorApiResponse,
	validatorsMap map[uint32][][]byte,
	currentList common.PeerType,
) {
	for shardID, shardValidators := range validatorsMap {
		for _, val := range shardValidators {
			encodedKey := vp.validatorPubKeyConverter.Encode(val)
			foundInTrieValidator, ok := newCache[encodedKey]
			peerType := string(currentList)

			if !ok || foundInTrieValidator == nil {
				newCache[encodedKey] = &state.ValidatorApiResponse{}
				newCache[encodedKey].ShardId = shardID
				newCache[encodedKey].ValidatorStatus = peerType
				log.Debug("validator from map not found in trie", "pk", encodedKey, "map", peerType)
				continue
			}

			trieList := common.PeerType(foundInTrieValidator.ValidatorStatus)
			if shouldCombine(trieList, currentList) {
				peerType = fmt.Sprintf(common.CombinedPeerType, currentList, trieList)
			}

			newCache[encodedKey].ShardId = shardID
			newCache[encodedKey].ValidatorStatus = peerType
		}
	}
}

func shouldCombine(triePeerType common.PeerType, currentPeerType common.PeerType) bool {
	// currently just "eligible (leaving)" or "waiting (leaving)" are allowed
	isLeaving := triePeerType == common.LeavingList
	isEligibleOrWaiting := currentPeerType == common.EligibleList ||
		currentPeerType == common.WaitingList

	return isLeaving && isEligibleOrWaiting
}

// IsInterfaceNil returns true if there is no value under the interface
func (vp *validatorsProvider) IsInterfaceNil() bool {
	return vp == nil
}

// Close - frees up everything, cancels long running methods
func (vp *validatorsProvider) Close() error {
	vp.cancelFunc()

	return nil
}
