package shard

import (
	"fmt"
	"io"
	"sort"

	arwen12 "github.com/ElrondNetwork/arwen-wasm-vm/arwen"
	arwenHost12 "github.com/ElrondNetwork/arwen-wasm-vm/arwen/host"
	ipcNodePart12 "github.com/ElrondNetwork/arwen-wasm-vm/ipc/nodepart"
	arwen13 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen"
	arwenHost13 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen/host"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config                         config.VirtualMachineConfig
	blockChainHookImpl             *hooks.BlockChainHookImpl
	cryptoHook                     vmcommon.CryptoHook
	blockGasLimit                  uint64
	gasSchedule                    core.GasScheduleNotifier
	builtinFunctions               vmcommon.FunctionNames
	epochNotifier                  process.EpochNotifier
	deployEnableEpoch              uint32
	aheadOfTimeGasUsageEnableEpoch uint32
	arwenV3EnableEpoch             uint32
	container                      process.VirtualMachinesContainer
	arwenVersions                  []config.ArwenVersionByEpoch
	arwenChangeLocker              process.Locker
}

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config                         config.VirtualMachineConfig
	BlockGasLimit                  uint64
	GasSchedule                    core.GasScheduleNotifier
	ArgBlockChainHook              hooks.ArgBlockChainHook
	EpochNotifier                  process.EpochNotifier
	DeployEnableEpoch              uint32
	AheadOfTimeGasUsageEnableEpoch uint32
	ArwenV3EnableEpoch             uint32
	ArwenChangeLocker              process.Locker
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(args ArgVMContainerFactory) (*vmContainerFactory, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNilReflect(args.ArwenChangeLocker) {
		return nil, process.ErrNilLocker
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(args.ArgBlockChainHook)
	if err != nil {
		return nil, err
	}

	cryptoHook := hooks.NewVMCryptoHook()
	builtinFunctions := blockChainHookImpl.GetBuiltinFunctionNames()

	vmf := &vmContainerFactory{
		config:                         args.Config,
		blockChainHookImpl:             blockChainHookImpl,
		cryptoHook:                     cryptoHook,
		blockGasLimit:                  args.BlockGasLimit,
		gasSchedule:                    args.GasSchedule,
		builtinFunctions:               builtinFunctions,
		epochNotifier:                  args.EpochNotifier,
		deployEnableEpoch:              args.DeployEnableEpoch,
		aheadOfTimeGasUsageEnableEpoch: args.AheadOfTimeGasUsageEnableEpoch,
		arwenV3EnableEpoch:             args.ArwenV3EnableEpoch,
		container:                      nil,
		arwenChangeLocker:              args.ArwenChangeLocker,
	}

	vmf.arwenVersions = args.Config.ArwenVersions
	vmf.sortArwenVersions()
	err = vmf.validateArwenVersions()
	if err != nil {
		return nil, err
	}

	return vmf, nil
}

func (vmf *vmContainerFactory) sortArwenVersions() {
	sort.Slice(vmf.arwenVersions, func(i, j int) bool {
		return vmf.arwenVersions[i].StartEpoch < vmf.arwenVersions[j].StartEpoch
	})
}

func (vmf *vmContainerFactory) validateArwenVersions() error {
	if len(vmf.arwenVersions) == 0 {
		return process.ErrEmptyVersionsByEpochsList
	}

	currentEpoch := uint32(0)
	for idx, ver := range vmf.arwenVersions {
		if idx == 0 && ver.StartEpoch != 0 {
			return fmt.Errorf("%w first version should start on epoch 0",
				process.ErrInvalidVersionOnEpochValues)
		}

		if idx > 0 && currentEpoch >= ver.StartEpoch {
			return fmt.Errorf("%w, StartEpoch is greater or equal to next epoch StartEpoch value, version %s",
				process.ErrInvalidVersionOnEpochValues, ver.Version)
		}
		currentEpoch = ver.StartEpoch

		if len(ver.Version) > core.MaxSoftwareVersionLengthInBytes {
			return fmt.Errorf("%w for version %s",
				process.ErrInvalidVersionStringTooLong, ver.Version)
		}
	}

	return nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	vmf.arwenChangeLocker.Lock()
	defer vmf.arwenChangeLocker.Unlock()

	version := vmf.getMatchingVersion(vmf.epochNotifier.CurrentEpoch())
	currentVM, err := vmf.createArwenVM(version)
	if err != nil {
		return nil, err
	}
	vmf.gasSchedule.RegisterNotifyHandler(currentVM)

	err = container.Add(factory.ArwenVirtualMachine, currentVM)
	if err != nil {
		return nil, err
	}

	// The vmContainerFactory keeps a reference to the container it has created,
	// in order to replace, from within the container, the VM instances that
	// become out-of-date after specific epochs.
	vmf.container = container
	vmf.epochNotifier.RegisterNotifyHandler(vmf)

	return container, nil
}

// EpochConfirmed updates the VM version in the container, depending on the epoch
func (vmf *vmContainerFactory) EpochConfirmed(epoch uint32) {
	vmf.ensureCorrectArwenVersion(epoch)
}

func (vmf *vmContainerFactory) ensureCorrectArwenVersion(epoch uint32) {
	newVersion := vmf.getMatchingVersion(epoch)
	currentArwenVM, err := vmf.container.Get(factory.ArwenVirtualMachine)
	if err != nil {
		logVMContainerFactory.Error("cannot retrieve Arwen VM from container", "epoch", epoch)
		return
	}

	logVMContainerFactory.Debug("ensureCorrectArwenVersion", "version now", currentArwenVM.GetVersion())

	if !vmf.shouldReplaceArwenInstance(newVersion, currentArwenVM) {
		return
	}

	vmf.arwenChangeLocker.Lock()
	defer vmf.arwenChangeLocker.Unlock()

	vmf.closePreviousVM(currentArwenVM)
	newArwenVM, err := vmf.createArwenVM(newVersion)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Arwen VM", "epoch", epoch)
		return
	}

	err = vmf.container.Replace(factory.ArwenVirtualMachine, newArwenVM)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Arwen VM", "epoch", epoch)
		return
	}

	logVMContainerFactory.Debug("Arwen VM replaced", "epoch", epoch, "version", newArwenVM.GetVersion())
}

func (vmf *vmContainerFactory) shouldReplaceArwenInstance(
	newVersion config.ArwenVersionByEpoch,
	currentVM vmcommon.VMExecutionHandler,
) bool {
	specificVersionRequired := newVersion.Version != "*"
	differentVersion := newVersion.Version != currentVM.GetVersion()

	return specificVersionRequired && differentVersion
}

func (vmf *vmContainerFactory) isArwenOutOfProcess(vm vmcommon.VMExecutionHandler) bool {
	_, ok := vm.(*ipcNodePart12.ArwenDriver)
	return ok
}

func (vmf *vmContainerFactory) createArwenVM(version config.ArwenVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	return vmf.createInProcessArwenVMByVersion(version)
}

func (vmf *vmContainerFactory) getMatchingVersion(epoch uint32) config.ArwenVersionByEpoch {
	matchingVersion := vmf.arwenVersions[len(vmf.arwenVersions)-1]
	for idx := 0; idx < len(vmf.arwenVersions)-1; idx++ {
		crtVer := vmf.arwenVersions[idx]
		nextVer := vmf.arwenVersions[idx+1]
		if crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch {
			return crtVer
		}
	}

	return matchingVersion
}

func (vmf *vmContainerFactory) createInProcessArwenVMByVersion(version config.ArwenVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessArwenVM", "version", version)
	switch version.Version {
	case "v1.2":
		return vmf.createInProcessArwenVMV12()
	default:
		return vmf.createInProcessArwenVMV13()
	}
}

func (vmf *vmContainerFactory) createInProcessArwenVMV12() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen12.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ArwenV2EnableEpoch:       vmf.deployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.aheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.deployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.arwenV3EnableEpoch,
	}
	return arwenHost12.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVMV13() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen13.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ArwenV2EnableEpoch:       vmf.deployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.aheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.deployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.arwenV3EnableEpoch,
	}
	return arwenHost13.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) closePreviousVM(vm vmcommon.VMExecutionHandler) {
	vmf.blockChainHookImpl.ClearCompiledCodes()
	logVMContainerFactory.Debug("AOT compilation cache cleared")

	vmAsCloser, ok := vm.(io.Closer)
	if ok {
		err := vmAsCloser.Close()
		logVMContainerFactory.LogIfError(err, "closePreviousVM", "error", err)
	}
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
