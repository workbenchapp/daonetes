// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smartwallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Creates a struct containing a reverse mapping of a subaccount to a
// [SmartWallet].
type CreateSubaccountInfo struct {
	Bump           *uint8
	Subaccount     *ag_solanago.PublicKey
	SmartWallet    *ag_solanago.PublicKey
	Index          *uint64
	SubaccountType *SubaccountType

	// [0] = [WRITE] subaccountInfo
	// ··········· The [SubaccountInfo] to create.
	//
	// [1] = [WRITE, SIGNER] payer
	// ··········· Payer to create the [SubaccountInfo].
	//
	// [2] = [] systemProgram
	// ··········· The [System] program.
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCreateSubaccountInfoInstructionBuilder creates a new `CreateSubaccountInfo` instruction builder.
func NewCreateSubaccountInfoInstructionBuilder() *CreateSubaccountInfo {
	nd := &CreateSubaccountInfo{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *CreateSubaccountInfo) SetBump(bump uint8) *CreateSubaccountInfo {
	inst.Bump = &bump
	return inst
}

// SetSubaccount sets the "subaccount" parameter.
func (inst *CreateSubaccountInfo) SetSubaccount(subaccount ag_solanago.PublicKey) *CreateSubaccountInfo {
	inst.Subaccount = &subaccount
	return inst
}

// SetSmartWallet sets the "smartWallet" parameter.
func (inst *CreateSubaccountInfo) SetSmartWallet(smartWallet ag_solanago.PublicKey) *CreateSubaccountInfo {
	inst.SmartWallet = &smartWallet
	return inst
}

// SetIndex sets the "index" parameter.
func (inst *CreateSubaccountInfo) SetIndex(index uint64) *CreateSubaccountInfo {
	inst.Index = &index
	return inst
}

// SetSubaccountType sets the "subaccountType" parameter.
func (inst *CreateSubaccountInfo) SetSubaccountType(subaccountType SubaccountType) *CreateSubaccountInfo {
	inst.SubaccountType = &subaccountType
	return inst
}

// SetSubaccountInfoAccount sets the "subaccountInfo" account.
// The [SubaccountInfo] to create.
func (inst *CreateSubaccountInfo) SetSubaccountInfoAccount(subaccountInfo ag_solanago.PublicKey) *CreateSubaccountInfo {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(subaccountInfo).WRITE()
	return inst
}

// GetSubaccountInfoAccount gets the "subaccountInfo" account.
// The [SubaccountInfo] to create.
func (inst *CreateSubaccountInfo) GetSubaccountInfoAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetPayerAccount sets the "payer" account.
// Payer to create the [SubaccountInfo].
func (inst *CreateSubaccountInfo) SetPayerAccount(payer ag_solanago.PublicKey) *CreateSubaccountInfo {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
// Payer to create the [SubaccountInfo].
func (inst *CreateSubaccountInfo) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetSystemProgramAccount sets the "systemProgram" account.
// The [System] program.
func (inst *CreateSubaccountInfo) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateSubaccountInfo {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
// The [System] program.
func (inst *CreateSubaccountInfo) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

func (inst CreateSubaccountInfo) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CreateSubaccountInfo,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateSubaccountInfo) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateSubaccountInfo) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
		if inst.Subaccount == nil {
			return errors.New("Subaccount parameter is not set")
		}
		if inst.SmartWallet == nil {
			return errors.New("SmartWallet parameter is not set")
		}
		if inst.Index == nil {
			return errors.New("Index parameter is not set")
		}
		if inst.SubaccountType == nil {
			return errors.New("SubaccountType parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SubaccountInfo is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CreateSubaccountInfo) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateSubaccountInfo")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=5]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("          Bump", *inst.Bump))
						paramsBranch.Child(ag_format.Param("    Subaccount", *inst.Subaccount))
						paramsBranch.Child(ag_format.Param("   SmartWallet", *inst.SmartWallet))
						paramsBranch.Child(ag_format.Param("         Index", *inst.Index))
						paramsBranch.Child(ag_format.Param("SubaccountType", *inst.SubaccountType))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=3]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("subaccountInfo", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("         payer", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta(" systemProgram", inst.AccountMetaSlice.Get(2)))
					})
				})
		})
}

func (obj CreateSubaccountInfo) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Subaccount` param:
	err = encoder.Encode(obj.Subaccount)
	if err != nil {
		return err
	}
	// Serialize `SmartWallet` param:
	err = encoder.Encode(obj.SmartWallet)
	if err != nil {
		return err
	}
	// Serialize `Index` param:
	err = encoder.Encode(obj.Index)
	if err != nil {
		return err
	}
	// Serialize `SubaccountType` param:
	err = encoder.Encode(obj.SubaccountType)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateSubaccountInfo) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Subaccount`:
	err = decoder.Decode(&obj.Subaccount)
	if err != nil {
		return err
	}
	// Deserialize `SmartWallet`:
	err = decoder.Decode(&obj.SmartWallet)
	if err != nil {
		return err
	}
	// Deserialize `Index`:
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}
	// Deserialize `SubaccountType`:
	err = decoder.Decode(&obj.SubaccountType)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateSubaccountInfoInstruction declares a new CreateSubaccountInfo instruction with the provided parameters and accounts.
func NewCreateSubaccountInfoInstruction(
	// Parameters:
	bump uint8,
	subaccount ag_solanago.PublicKey,
	smartWallet ag_solanago.PublicKey,
	index uint64,
	subaccountType SubaccountType,
	// Accounts:
	subaccountInfo ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CreateSubaccountInfo {
	return NewCreateSubaccountInfoInstructionBuilder().
		SetBump(bump).
		SetSubaccount(subaccount).
		SetSmartWallet(smartWallet).
		SetIndex(index).
		SetSubaccountType(subaccountType).
		SetSubaccountInfoAccount(subaccountInfo).
		SetPayerAccount(payer).
		SetSystemProgramAccount(systemProgram)
}
