package keysManagement

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// ArgsKeysHandler is the argument DTO struct for the NewKeysHandler constructor function
type ArgsKeysHandler struct {
	ManagedPeersHolder ManagedPeersHolder
	PrivateKey         crypto.PrivateKey
	Pid                core.PeerID
}

// keysHandler will manage all keys available on the node either in single signer mode or multi key mode
type keysHandler struct {
	managedPeersHolder ManagedPeersHolder
	privateKey         crypto.PrivateKey
	publicKey          crypto.PublicKey
	publicKeyBytes     []byte
	pid                core.PeerID
}

// NewKeysHandler will create a new instance of type keysHandler
func NewKeysHandler(args ArgsKeysHandler) (*keysHandler, error) {
	err := checkArgsKeysHandler(args)
	if err != nil {
		return nil, err
	}

	pk := args.PrivateKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	return &keysHandler{
		managedPeersHolder: args.ManagedPeersHolder,
		privateKey:         args.PrivateKey,
		publicKey:          pk,
		publicKeyBytes:     pkBytes,
		pid:                args.Pid,
	}, nil
}

func checkArgsKeysHandler(args ArgsKeysHandler) error {
	if check.IfNil(args.ManagedPeersHolder) {
		return errNilManagedPeersHolder
	}
	if check.IfNil(args.PrivateKey) {
		return errNilPrivateKey
	}
	if len(args.Pid) == 0 {
		return errEmptyPeerID
	}

	return nil
}

// GetHandledPrivateKey will return the correct private key by using the provided pkBytes to select from a stored one
// Returns the current private key if the pkBytes is not handled by the current node
func (handler *keysHandler) GetHandledPrivateKey(pkBytes []byte) crypto.PrivateKey {
	if handler.IsOriginalPublicKeyOfTheNode(pkBytes) {
		return handler.privateKey
	}

	privateKey, err := handler.managedPeersHolder.GetPrivateKey(pkBytes)
	if err != nil {
		log.Warn("setup error in keysHandler.GetHandledPrivateKey, returning original private key", "error", err)

		return handler.privateKey
	}

	return privateKey
}

// GetP2PIdentity returns the associated p2p identity with the provided public key bytes: the private key and the peer ID
func (handler *keysHandler) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	return handler.managedPeersHolder.GetP2PIdentity(pkBytes)
}

// IsKeyManagedByCurrentNode will return if the provided key is a managed one and the current node should use it
func (handler *keysHandler) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	return handler.managedPeersHolder.IsKeyManagedByCurrentNode(pkBytes)
}

// IncrementRoundsWithoutReceivedMessages increments the provided rounds without received messages counter on the provided public key
func (handler *keysHandler) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	handler.managedPeersHolder.IncrementRoundsWithoutReceivedMessages(pkBytes)
}

// GetAssociatedPid will return the associated peer ID from the provided public key bytes. Will search in the managed keys mapping
// if the public key is not the original public key of the node
func (handler *keysHandler) GetAssociatedPid(pkBytes []byte) core.PeerID {
	if handler.IsOriginalPublicKeyOfTheNode(pkBytes) {
		return handler.pid
	}

	_, pid, err := handler.managedPeersHolder.GetP2PIdentity(pkBytes)
	if err != nil {
		log.Warn("setup error in keysHandler.GetAssociatedPid, returning original pid", "error", err)

		return handler.pid
	}

	return pid
}

// IsOriginalPublicKeyOfTheNode returns true if the provided public key bytes are the original ones used by the node
func (handler *keysHandler) IsOriginalPublicKeyOfTheNode(pkBytes []byte) bool {
	return bytes.Equal(pkBytes, handler.publicKeyBytes)
}

// UpdatePublicKeyLiveness update the provided public key liveness if the provided pid is not managed by the current node
func (handler *keysHandler) UpdatePublicKeyLiveness(pkBytes []byte, pid core.PeerID) {
	if bytes.Equal(handler.pid.Bytes(), pid.Bytes()) {
		return
	}

	handler.managedPeersHolder.ResetRoundsWithoutReceivedMessages(pkBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *keysHandler) IsInterfaceNil() bool {
	return handler == nil
}
