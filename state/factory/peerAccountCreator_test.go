package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
)

func TestPeerAccountCreator_CreateAccountNilAddress(t *testing.T) {
	t.Parallel()

	accF := factory.NewPeerAccountCreator()

	_, ok := accF.(*factory.PeerAccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(nil, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

func TestPeerAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF := factory.NewPeerAccountCreator()
	assert.False(t, check.IfNil(accF))

	_, ok := accF.(*factory.PeerAccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(make([]byte, 32), &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	assert.NotNil(t, acc)
	assert.Nil(t, err)
}
