// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smartwallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Creates a new [Transaction] account with time delay.
type CreateTransactionWithTimelock struct {
	Bump         *uint8
	Instructions *[]TXInstruction
	Eta          *int64

	// [0] = [WRITE] smartWallet
	// ··········· The [SmartWallet].
	//
	// [1] = [WRITE] transaction
	// ··········· The [Transaction].
	//
	// [2] = [SIGNER] proposer
	// ··········· One of the owners. Checked in the handler via [SmartWallet::try_owner_index].
	//
	// [3] = [WRITE, SIGNER] payer
	// ··········· Payer to create the [Transaction].
	//
	// [4] = [] systemProgram
	// ··········· The [System] program.
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCreateTransactionWithTimelockInstructionBuilder creates a new `CreateTransactionWithTimelock` instruction builder.
func NewCreateTransactionWithTimelockInstructionBuilder() *CreateTransactionWithTimelock {
	nd := &CreateTransactionWithTimelock{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *CreateTransactionWithTimelock) SetBump(bump uint8) *CreateTransactionWithTimelock {
	inst.Bump = &bump
	return inst
}

// SetInstructions sets the "instructions" parameter.
func (inst *CreateTransactionWithTimelock) SetInstructions(instructions []TXInstruction) *CreateTransactionWithTimelock {
	inst.Instructions = &instructions
	return inst
}

// SetEta sets the "eta" parameter.
func (inst *CreateTransactionWithTimelock) SetEta(eta int64) *CreateTransactionWithTimelock {
	inst.Eta = &eta
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
// The [SmartWallet].
func (inst *CreateTransactionWithTimelock) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
// The [SmartWallet].
func (inst *CreateTransactionWithTimelock) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetTransactionAccount sets the "transaction" account.
// The [Transaction].
func (inst *CreateTransactionWithTimelock) SetTransactionAccount(transaction ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(transaction).WRITE()
	return inst
}

// GetTransactionAccount gets the "transaction" account.
// The [Transaction].
func (inst *CreateTransactionWithTimelock) GetTransactionAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetProposerAccount sets the "proposer" account.
// One of the owners. Checked in the handler via [SmartWallet::try_owner_index].
func (inst *CreateTransactionWithTimelock) SetProposerAccount(proposer ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(proposer).SIGNER()
	return inst
}

// GetProposerAccount gets the "proposer" account.
// One of the owners. Checked in the handler via [SmartWallet::try_owner_index].
func (inst *CreateTransactionWithTimelock) GetProposerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetPayerAccount sets the "payer" account.
// Payer to create the [Transaction].
func (inst *CreateTransactionWithTimelock) SetPayerAccount(payer ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
// Payer to create the [Transaction].
func (inst *CreateTransactionWithTimelock) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetSystemProgramAccount sets the "systemProgram" account.
// The [System] program.
func (inst *CreateTransactionWithTimelock) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
// The [System] program.
func (inst *CreateTransactionWithTimelock) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

func (inst CreateTransactionWithTimelock) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CreateTransactionWithTimelock,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateTransactionWithTimelock) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateTransactionWithTimelock) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
		if inst.Instructions == nil {
			return errors.New("Instructions parameter is not set")
		}
		if inst.Eta == nil {
			return errors.New("Eta parameter is not set")
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
			return errors.New("accounts.Proposer is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CreateTransactionWithTimelock) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateTransactionWithTimelock")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=3]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("        Bump", *inst.Bump))
						paramsBranch.Child(ag_format.Param("Instructions", *inst.Instructions))
						paramsBranch.Child(ag_format.Param("         Eta", *inst.Eta))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  smartWallet", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("  transaction", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("     proposer", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("        payer", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

func (obj CreateTransactionWithTimelock) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Instructions` param:
	err = encoder.Encode(obj.Instructions)
	if err != nil {
		return err
	}
	// Serialize `Eta` param:
	err = encoder.Encode(obj.Eta)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateTransactionWithTimelock) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Instructions`:
	err = decoder.Decode(&obj.Instructions)
	if err != nil {
		return err
	}
	// Deserialize `Eta`:
	err = decoder.Decode(&obj.Eta)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateTransactionWithTimelockInstruction declares a new CreateTransactionWithTimelock instruction with the provided parameters and accounts.
func NewCreateTransactionWithTimelockInstruction(
	// Parameters:
	bump uint8,
	instructions []TXInstruction,
	eta int64,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	transaction ag_solanago.PublicKey,
	proposer ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CreateTransactionWithTimelock {
	return NewCreateTransactionWithTimelockInstructionBuilder().
		SetBump(bump).
		SetInstructions(instructions).
		SetEta(eta).
		SetSmartWalletAccount(smartWallet).
		SetTransactionAccount(transaction).
		SetProposerAccount(proposer).
		SetPayerAccount(payer).
		SetSystemProgramAccount(systemProgram)
}
