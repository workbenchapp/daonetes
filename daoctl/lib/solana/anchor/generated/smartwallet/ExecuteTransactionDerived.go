// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smartwallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Executes the given transaction signed by the given derived address,
// if threshold owners have signed it.
// This allows a Smart Wallet to receive SOL.
type ExecuteTransactionDerived struct {
	Index *uint64
	Bump  *uint8

	// [0] = [] smartWallet
	// ··········· The [SmartWallet].
	//
	// [1] = [WRITE] transaction
	// ··········· The [Transaction] to execute.
	//
	// [2] = [SIGNER] owner
	// ··········· An owner of the [SmartWallet].
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewExecuteTransactionDerivedInstructionBuilder creates a new `ExecuteTransactionDerived` instruction builder.
func NewExecuteTransactionDerivedInstructionBuilder() *ExecuteTransactionDerived {
	nd := &ExecuteTransactionDerived{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

// SetIndex sets the "index" parameter.
func (inst *ExecuteTransactionDerived) SetIndex(index uint64) *ExecuteTransactionDerived {
	inst.Index = &index
	return inst
}

// SetBump sets the "bump" parameter.
func (inst *ExecuteTransactionDerived) SetBump(bump uint8) *ExecuteTransactionDerived {
	inst.Bump = &bump
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
// The [SmartWallet].
func (inst *ExecuteTransactionDerived) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *ExecuteTransactionDerived {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet)
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
// The [SmartWallet].
func (inst *ExecuteTransactionDerived) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetTransactionAccount sets the "transaction" account.
// The [Transaction] to execute.
func (inst *ExecuteTransactionDerived) SetTransactionAccount(transaction ag_solanago.PublicKey) *ExecuteTransactionDerived {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(transaction).WRITE()
	return inst
}

// GetTransactionAccount gets the "transaction" account.
// The [Transaction] to execute.
func (inst *ExecuteTransactionDerived) GetTransactionAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetOwnerAccount sets the "owner" account.
// An owner of the [SmartWallet].
func (inst *ExecuteTransactionDerived) SetOwnerAccount(owner ag_solanago.PublicKey) *ExecuteTransactionDerived {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
// An owner of the [SmartWallet].
func (inst *ExecuteTransactionDerived) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

func (inst ExecuteTransactionDerived) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_ExecuteTransactionDerived,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst ExecuteTransactionDerived) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ExecuteTransactionDerived) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Index == nil {
			return errors.New("Index parameter is not set")
		}
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SmartWallet is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Transaction is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Owner is not set")
		}
	}
	return nil
}

func (inst *ExecuteTransactionDerived) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ExecuteTransactionDerived")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=2]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Index", *inst.Index))
						paramsBranch.Child(ag_format.Param(" Bump", *inst.Bump))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=3]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("smartWallet", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("transaction", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("      owner", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

func (obj ExecuteTransactionDerived) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
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
	return nil
}
func (obj *ExecuteTransactionDerived) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
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
	return nil
}

// NewExecuteTransactionDerivedInstruction declares a new ExecuteTransactionDerived instruction with the provided parameters and accounts.
func NewExecuteTransactionDerivedInstruction(
	// Parameters:
	index uint64,
	bump uint8,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	transaction ag_solanago.PublicKey,
	owner ag_solanago.PublicKey) *ExecuteTransactionDerived {
	return NewExecuteTransactionDerivedInstructionBuilder().
		SetIndex(index).
		SetBump(bump).
		SetSmartWalletAccount(smartWallet).
		SetTransactionAccount(transaction).
		SetOwnerAccount(owner)
}
