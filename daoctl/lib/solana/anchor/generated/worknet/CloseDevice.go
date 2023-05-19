// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package worknet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CloseDevice is the `closeDevice` instruction.
type CloseDevice struct {

	// [0] = [WRITE, SIGNER] groupAuthority
	//
	// [1] = [WRITE] device
	//
	// [2] = [WRITE] workGroup
	//
	// [3] = [] systemProgram
	//
	// [4] = [] tokenProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCloseDeviceInstructionBuilder creates a new `CloseDevice` instruction builder.
func NewCloseDeviceInstructionBuilder() *CloseDevice {
	nd := &CloseDevice{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetGroupAuthorityAccount sets the "groupAuthority" account.
func (inst *CloseDevice) SetGroupAuthorityAccount(groupAuthority ag_solanago.PublicKey) *CloseDevice {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(groupAuthority).WRITE().SIGNER()
	return inst
}

// GetGroupAuthorityAccount gets the "groupAuthority" account.
func (inst *CloseDevice) GetGroupAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetDeviceAccount sets the "device" account.
func (inst *CloseDevice) SetDeviceAccount(device ag_solanago.PublicKey) *CloseDevice {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(device).WRITE()
	return inst
}

// GetDeviceAccount gets the "device" account.
func (inst *CloseDevice) GetDeviceAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetWorkGroupAccount sets the "workGroup" account.
func (inst *CloseDevice) SetWorkGroupAccount(workGroup ag_solanago.PublicKey) *CloseDevice {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(workGroup).WRITE()
	return inst
}

// GetWorkGroupAccount gets the "workGroup" account.
func (inst *CloseDevice) GetWorkGroupAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CloseDevice) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CloseDevice {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CloseDevice) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *CloseDevice) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *CloseDevice {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *CloseDevice) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

func (inst CloseDevice) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CloseDevice,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CloseDevice) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CloseDevice) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.GroupAuthority is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Device is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.WorkGroup is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
	}
	return nil
}

func (inst *CloseDevice) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CloseDevice")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("groupAuthority", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("        device", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("     workGroup", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta(" systemProgram", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("  tokenProgram", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

func (obj CloseDevice) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *CloseDevice) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewCloseDeviceInstruction declares a new CloseDevice instruction with the provided parameters and accounts.
func NewCloseDeviceInstruction(
	// Accounts:
	groupAuthority ag_solanago.PublicKey,
	device ag_solanago.PublicKey,
	workGroup ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey) *CloseDevice {
	return NewCloseDeviceInstructionBuilder().
		SetGroupAuthorityAccount(groupAuthority).
		SetDeviceAccount(device).
		SetWorkGroupAccount(workGroup).
		SetSystemProgramAccount(systemProgram).
		SetTokenProgramAccount(tokenProgram)
}
