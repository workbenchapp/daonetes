// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smartwallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Invokes an arbitrary instruction as a PDA derived from the owner,
// i.e. as an "Owner Invoker".
//
// This is useful for using the multisig as a whitelist or as a council,
// e.g. a whitelist of approved owners.
//
// # Arguments
// - `index` - The index of the owner-invoker.
// - `bump` - Bump seed of the owner-invoker.
// - `invoker` - The owner-invoker.
// - `data` - The raw bytes of the instruction data.
type OwnerInvokeInstructionV2 struct {
	Index   *uint64
	Bump    *uint8
	Invoker *ag_solanago.PublicKey
	Data    *[]byte

	// [0] = [] smartWallet
	// ··········· The [SmartWallet].
	//
	// [1] = [SIGNER] owner
	// ··········· An owner of the [SmartWallet].
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewOwnerInvokeInstructionV2InstructionBuilder creates a new `OwnerInvokeInstructionV2` instruction builder.
func NewOwnerInvokeInstructionV2InstructionBuilder() *OwnerInvokeInstructionV2 {
	nd := &OwnerInvokeInstructionV2{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetIndex sets the "index" parameter.
func (inst *OwnerInvokeInstructionV2) SetIndex(index uint64) *OwnerInvokeInstructionV2 {
	inst.Index = &index
	return inst
}

// SetBump sets the "bump" parameter.
func (inst *OwnerInvokeInstructionV2) SetBump(bump uint8) *OwnerInvokeInstructionV2 {
	inst.Bump = &bump
	return inst
}

// SetInvoker sets the "invoker" parameter.
func (inst *OwnerInvokeInstructionV2) SetInvoker(invoker ag_solanago.PublicKey) *OwnerInvokeInstructionV2 {
	inst.Invoker = &invoker
	return inst
}

// SetData sets the "data" parameter.
func (inst *OwnerInvokeInstructionV2) SetData(data []byte) *OwnerInvokeInstructionV2 {
	inst.Data = &data
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
// The [SmartWallet].
func (inst *OwnerInvokeInstructionV2) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *OwnerInvokeInstructionV2 {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet)
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
// The [SmartWallet].
func (inst *OwnerInvokeInstructionV2) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOwnerAccount sets the "owner" account.
// An owner of the [SmartWallet].
func (inst *OwnerInvokeInstructionV2) SetOwnerAccount(owner ag_solanago.PublicKey) *OwnerInvokeInstructionV2 {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
// An owner of the [SmartWallet].
func (inst *OwnerInvokeInstructionV2) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst OwnerInvokeInstructionV2) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_OwnerInvokeInstructionV2,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst OwnerInvokeInstructionV2) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *OwnerInvokeInstructionV2) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Index == nil {
			return errors.New("Index parameter is not set")
		}
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
		if inst.Invoker == nil {
			return errors.New("Invoker parameter is not set")
		}
		if inst.Data == nil {
			return errors.New("Data parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SmartWallet is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Owner is not set")
		}
	}
	return nil
}

func (inst *OwnerInvokeInstructionV2) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("OwnerInvokeInstructionV2")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=4]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("  Index", *inst.Index))
						paramsBranch.Child(ag_format.Param("   Bump", *inst.Bump))
						paramsBranch.Child(ag_format.Param("Invoker", *inst.Invoker))
						paramsBranch.Child(ag_format.Param("   Data", *inst.Data))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("smartWallet", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("      owner", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj OwnerInvokeInstructionV2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Index` param:
	err = encoder.Encode(obj.Index)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Invoker` param:
	err = encoder.Encode(obj.Invoker)
	if err != nil {
		return err
	}
	// Serialize `Data` param:
	err = encoder.Encode(obj.Data)
	if err != nil {
		return err
	}
	return nil
}
func (obj *OwnerInvokeInstructionV2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Index`:
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Invoker`:
	err = decoder.Decode(&obj.Invoker)
	if err != nil {
		return err
	}
	// Deserialize `Data`:
	err = decoder.Decode(&obj.Data)
	if err != nil {
		return err
	}
	return nil
}

// NewOwnerInvokeInstructionV2Instruction declares a new OwnerInvokeInstructionV2 instruction with the provided parameters and accounts.
func NewOwnerInvokeInstructionV2Instruction(
	// Parameters:
	index uint64,
	bump uint8,
	invoker ag_solanago.PublicKey,
	data []byte,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	owner ag_solanago.PublicKey) *OwnerInvokeInstructionV2 {
	return NewOwnerInvokeInstructionV2InstructionBuilder().
		SetIndex(index).
		SetBump(bump).
		SetInvoker(invoker).
		SetData(data).
		SetSmartWalletAccount(smartWallet).
		SetOwnerAccount(owner)
}
