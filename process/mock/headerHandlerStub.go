package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// HeaderHandlerStub -
type HeaderHandlerStub struct {
	GetMiniBlockHeadersWithDstCalled       func(destId uint32) map[string]uint32
	GetOrderedCrossMiniblocksWithDstCalled func(destId uint32) []*data.MiniBlockInfo
	GetPubKeysBitmapCalled                 func() []byte
	GetSignatureCalled                     func() []byte
	GetRootHashCalled                      func() []byte
	GetRandSeedCalled                      func() []byte
	GetPrevRandSeedCalled                  func() []byte
	GetPrevHashCalled                      func() []byte
	CloneCalled                            func() data.HeaderHandler
	GetChainIDCalled                       func() []byte
	CheckChainIDCalled                     func(reference []byte) error
	GetAccumulatedFeesCalled               func() *big.Int
	GetDeveloperFeesCalled                 func() *big.Int
	GetReservedCalled                      func() []byte
}

// GetAccumulatedFees -
func (hhs *HeaderHandlerStub) GetAccumulatedFees() *big.Int {
	if hhs.GetAccumulatedFeesCalled != nil {
		return hhs.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// GetDeveloperFees -
func (hhs *HeaderHandlerStub) GetDeveloperFees() *big.Int {
	if hhs.GetDeveloperFeesCalled != nil {
		return hhs.GetDeveloperFeesCalled()
	}
	return big.NewInt(0)
}

// SetAccumulatedFees -
func (hhs *HeaderHandlerStub) SetAccumulatedFees(_ *big.Int) {
	panic("implement me")
}

// SetDeveloperFees -
func (hhs *HeaderHandlerStub) SetDeveloperFees(_ *big.Int) {
	panic("implement me")
}

// SetDevFeesInEpoch -
func (hhs *HeaderHandlerStub) SetDevFeesInEpoch(value *big.Int){
	panic("implement me")
}

// GetReceiptsHash -
func (hhs *HeaderHandlerStub) GetReceiptsHash() []byte {
	return []byte("hash")
}

// Clone -
func (hhs *HeaderHandlerStub) ShallowClone() data.HeaderHandler {
	return hhs.CloneCalled()
}

// IsStartOfEpochBlock -
func (hhs *HeaderHandlerStub) IsStartOfEpochBlock() bool {
	return false
}

// GetShardID -
func (hhs *HeaderHandlerStub) GetShardID() uint32 {
	return 1
}

// SetShardID -
func (hhs *HeaderHandlerStub) SetShardID(_ uint32) {
}

// GetNonce -
func (hhs *HeaderHandlerStub) GetNonce() uint64 {
	return 1
}

// GetEpoch -
func (hhs *HeaderHandlerStub) GetEpoch() uint32 {
	panic("implement me")
}

// GetRound -
func (hhs *HeaderHandlerStub) GetRound() uint64 {
	return 1
}

// GetTimeStamp -
func (hhs *HeaderHandlerStub) GetTimeStamp() uint64 {
	panic("implement me")
}

// GetRootHash -
func (hhs *HeaderHandlerStub) GetRootHash() []byte {
	return hhs.GetRootHashCalled()
}

// GetPrevHash -
func (hhs *HeaderHandlerStub) GetPrevHash() []byte {
	return hhs.GetPrevHashCalled()
}

// GetPrevRandSeed -
func (hhs *HeaderHandlerStub) GetPrevRandSeed() []byte {
	return hhs.GetPrevRandSeedCalled()
}

// GetRandSeed -
func (hhs *HeaderHandlerStub) GetRandSeed() []byte {
	return hhs.GetRandSeedCalled()
}

// GetPubKeysBitmap -
func (hhs *HeaderHandlerStub) GetPubKeysBitmap() []byte {
	return hhs.GetPubKeysBitmapCalled()
}

// GetSignature -
func (hhs *HeaderHandlerStub) GetSignature() []byte {
	return hhs.GetSignatureCalled()
}

// GetLeaderSignature -
func (hhs *HeaderHandlerStub) GetLeaderSignature() []byte {
	return hhs.GetSignatureCalled()
}

// GetChainID -
func (hhs *HeaderHandlerStub) GetChainID() []byte {
	return hhs.GetChainIDCalled()
}

// GetTxCount -
func (hhs *HeaderHandlerStub) GetTxCount() uint32 {
	panic("implement me")
}

// GetReserved -
func (hhs *HeaderHandlerStub) GetReserved() []byte {
	if hhs.GetReservedCalled != nil {
		return hhs.GetReservedCalled()
	}

	return nil
}

// SetNonce -
func (hhs *HeaderHandlerStub) SetNonce(_ uint64) {
	panic("implement me")
}

// SetEpoch -
func (hhs *HeaderHandlerStub) SetEpoch(_ uint32) {
	panic("implement me")
}

// SetRound -
func (hhs *HeaderHandlerStub) SetRound(_ uint64) {
	panic("implement me")
}

// SetTimeStamp -
func (hhs *HeaderHandlerStub) SetTimeStamp(_ uint64) {
	panic("implement me")
}

// SetRootHash -
func (hhs *HeaderHandlerStub) SetRootHash(_ []byte) {
	panic("implement me")
}

// SetPrevHash -
func (hhs *HeaderHandlerStub) SetPrevHash(_ []byte) {
	panic("implement me")
}

// SetPrevRandSeed -
func (hhs *HeaderHandlerStub) SetPrevRandSeed(_ []byte) {
	panic("implement me")
}

// SetRandSeed -
func (hhs *HeaderHandlerStub) SetRandSeed(_ []byte) {
	panic("implement me")
}

// SetPubKeysBitmap -
func (hhs *HeaderHandlerStub) SetPubKeysBitmap(_ []byte) {
	panic("implement me")
}

// SetSignature -
func (hhs *HeaderHandlerStub) SetSignature(_ []byte) {
	panic("implement me")
}

// SetLeaderSignature -
func (hhs *HeaderHandlerStub) SetLeaderSignature(_ []byte) {
	panic("implement me")
}

// SetChainID -
func (hhs *HeaderHandlerStub) SetChainID(_ []byte) {
	panic("implement me")
}

// SetTxCount -
func (hhs *HeaderHandlerStub) SetTxCount(_ uint32) {
	panic("implement me")
}

// GetMiniBlockHeadersWithDst -
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	return hhs.GetMiniBlockHeadersWithDstCalled(destId)
}

// GetOrderedCrossMiniblocksWithDst -
func (hhs *HeaderHandlerStub) GetOrderedCrossMiniblocksWithDst(destId uint32) []*data.MiniBlockInfo {
	return hhs.GetOrderedCrossMiniblocksWithDstCalled(destId)
}

// GetMiniBlockHeadersHashes -
func (hhs *HeaderHandlerStub) GetMiniBlockHeadersHashes() [][]byte {
	panic("implement me")
}

// GetMiniBlockHeaders -
func (hhs *HeaderHandlerStub) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler{
	panic("implement me")
}

// GetMetaBlockHashes -
func (hhs *HeaderHandlerStub) GetMetaBlockHashes() [][]byte {
	panic("implement me")
}

// GetBlockBodyType -
func (hhs *HeaderHandlerStub) GetBlockBodyTypeInt32() int32 {
	panic("implement me")
}

// GetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) GetValidatorStatsRootHash() []byte {
	panic("implement me")
}

// SetValidatorStatsRootHash -
func (hhs *HeaderHandlerStub) SetValidatorStatsRootHash(_ []byte) {
	panic("implement me")
}

// SetMiniBlockHeaderHandlers -
func (hhs *HeaderHandlerStub) SetMiniBlockHeaderHandlers(_ []data.MiniBlockHeaderHandler) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hhs *HeaderHandlerStub) IsInterfaceNil() bool {
	return hhs == nil
}

// GetEpochStartMetaHash -
func (hhs *HeaderHandlerStub) GetEpochStartMetaHash() []byte {
	panic("implement me")
}

// GetSoftwareVersion -
func (hhs *HeaderHandlerStub) GetSoftwareVersion() []byte {
	return []byte("softwareVersion")
}

// SetSoftwareVersion -
func (hhs *HeaderHandlerStub) SetSoftwareVersion(_ []byte) {
}

// SetReceiptsHash -
func (hhs *HeaderHandlerStub) SetReceiptsHash(_ []byte) {
}

// SetMetaBlockHashes -
func (hhs *HeaderHandlerStub) SetMetaBlockHashes(_ [][]byte) {
}

// GetShardInfoHandlers -
func (hhs *HeaderHandlerStub) GetShardInfoHandlers() []data.ShardDataHandler{
	panic("implement me")
}

// GetEpochStartHandler -
func (hhs *HeaderHandlerStub) GetEpochStartHandler() data.EpochStartHandler {
	panic("implement me")
}

// GetDevFeesInEpoch -
func (hhs *HeaderHandlerStub) GetDevFeesInEpoch() *big.Int {
	panic("implement me")
}

// SetShardInfoHandlers -
func (hhs *HeaderHandlerStub) SetShardInfoHandlers(_ []data.ShardDataHandler) {
	panic("implement me")
}

// SetAccumulatedFeesInEpoch -
func (hhs *HeaderHandlerStub) SetAccumulatedFeesInEpoch(_ *big.Int) {
	panic("implement me")
}
