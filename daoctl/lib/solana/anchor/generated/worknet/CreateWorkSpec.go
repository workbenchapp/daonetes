// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package worknet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CreateWorkSpec is the `createWorkSpec` instruction.
type CreateWorkSpec struct {
	SpecName       *string
	WorkType       *WorkType
	UrlOrContents  *string
	ContentsSha256 *string
	MetadataUrl    *string
	Mutable        *bool

	// [0] = [WRITE, SIGNER] groupAuthority
	//
	// [1] = [WRITE] spec
	//
	// [2] = [WRITE] workGroup
	//
	// [3] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCreateWorkSpecInstructionBuilder creates a new `CreateWorkSpec` instruction builder.
func NewCreateWorkSpecInstructionBuilder() *CreateWorkSpec {
	nd := &CreateWorkSpec{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 4),
	}
	return nd
}

// SetSpecName sets the "specName" parameter.
func (inst *CreateWorkSpec) SetSpecName(specName string) *CreateWorkSpec {
	inst.SpecName = &specName
	return inst
}

// SetWorkType sets the "workType" parameter.
func (inst *CreateWorkSpec) SetWorkType(workType WorkType) *CreateWorkSpec {
	inst.WorkType = &workType
	return inst
}

// SetUrlOrContents sets the "urlOrContents" parameter.
func (inst *CreateWorkSpec) SetUrlOrContents(urlOrContents string) *CreateWorkSpec {
	inst.UrlOrContents = &urlOrContents
	return inst
}

// SetContentsSha256 sets the "contentsSha256" parameter.
func (inst *CreateWorkSpec) SetContentsSha256(contentsSha256 string) *CreateWorkSpec {
	inst.ContentsSha256 = &contentsSha256
	return inst
}

// SetMetadataUrl sets the "metadataUrl" parameter.
func (inst *CreateWorkSpec) SetMetadataUrl(metadataUrl string) *CreateWorkSpec {
	inst.MetadataUrl = &metadataUrl
	return inst
}

// SetMutable sets the "mutable" parameter.
func (inst *CreateWorkSpec) SetMutable(mutable bool) *CreateWorkSpec {
	inst.Mutable = &mutable
	return inst
}

// SetGroupAuthorityAccount sets the "groupAuthority" account.
func (inst *CreateWorkSpec) SetGroupAuthorityAccount(groupAuthority ag_solanago.PublicKey) *CreateWorkSpec {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(groupAuthority).WRITE().SIGNER()
	return inst
}

// GetGroupAuthorityAccount gets the "groupAuthority" account.
func (inst *CreateWorkSpec) GetGroupAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetSpecAccount sets the "spec" account.
func (inst *CreateWorkSpec) SetSpecAccount(spec ag_solanago.PublicKey) *CreateWorkSpec {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(spec).WRITE()
	return inst
}

// GetSpecAccount gets the "spec" account.
func (inst *CreateWorkSpec) GetSpecAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetWorkGroupAccount sets the "workGroup" account.
func (inst *CreateWorkSpec) SetWorkGroupAccount(workGroup ag_solanago.PublicKey) *CreateWorkSpec {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(workGroup).WRITE()
	return inst
}

// GetWorkGroupAccount gets the "workGroup" account.
func (inst *CreateWorkSpec) GetWorkGroupAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CreateWorkSpec) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateWorkSpec {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CreateWorkSpec) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

func (inst CreateWorkSpec) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CreateWorkSpec,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateWorkSpec) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateWorkSpec) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.SpecName == nil {
			return errors.New("SpecName parameter is not set")
		}
		if inst.WorkType == nil {
			return errors.New("WorkType parameter is not set")
		}
		if inst.UrlOrContents == nil {
			return errors.New("UrlOrContents parameter is not set")
		}
		if inst.ContentsSha256 == nil {
			return errors.New("ContentsSha256 parameter is not set")
		}
		if inst.MetadataUrl == nil {
			return errors.New("MetadataUrl parameter is not set")
		}
		if inst.Mutable == nil {
			return errors.New("Mutable parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.GroupAuthority is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Spec is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.WorkGroup is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CreateWorkSpec) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateWorkSpec")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=6]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("      SpecName", *inst.SpecName))
						paramsBranch.Child(ag_format.Param("      WorkType", *inst.WorkType))
						paramsBranch.Child(ag_format.Param(" UrlOrContents", *inst.UrlOrContents))
						paramsBranch.Child(ag_format.Param("ContentsSha256", *inst.ContentsSha256))
						paramsBranch.Child(ag_format.Param("   MetadataUrl", *inst.MetadataUrl))
						paramsBranch.Child(ag_format.Param("       Mutable", *inst.Mutable))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=4]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("groupAuthority", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("          spec", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("     workGroup", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta(" systemProgram", inst.AccountMetaSlice.Get(3)))
					})
				})
		})
}

func (obj CreateWorkSpec) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `SpecName` param:
	err = encoder.Encode(obj.SpecName)
	if err != nil {
		return err
	}
	// Serialize `WorkType` param:
	err = encoder.Encode(obj.WorkType)
	if err != nil {
		return err
	}
	// Serialize `UrlOrContents` param:
	err = encoder.Encode(obj.UrlOrContents)
	if err != nil {
		return err
	}
	// Serialize `ContentsSha256` param:
	err = encoder.Encode(obj.ContentsSha256)
	if err != nil {
		return err
	}
	// Serialize `MetadataUrl` param:
	err = encoder.Encode(obj.MetadataUrl)
	if err != nil {
		return err
	}
	// Serialize `Mutable` param:
	err = encoder.Encode(obj.Mutable)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateWorkSpec) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `SpecName`:
	err = decoder.Decode(&obj.SpecName)
	if err != nil {
		return err
	}
	// Deserialize `WorkType`:
	err = decoder.Decode(&obj.WorkType)
	if err != nil {
		return err
	}
	// Deserialize `UrlOrContents`:
	err = decoder.Decode(&obj.UrlOrContents)
	if err != nil {
		return err
	}
	// Deserialize `ContentsSha256`:
	err = decoder.Decode(&obj.ContentsSha256)
	if err != nil {
		return err
	}
	// Deserialize `MetadataUrl`:
	err = decoder.Decode(&obj.MetadataUrl)
	if err != nil {
		return err
	}
	// Deserialize `Mutable`:
	err = decoder.Decode(&obj.Mutable)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateWorkSpecInstruction declares a new CreateWorkSpec instruction with the provided parameters and accounts.
func NewCreateWorkSpecInstruction(
	// Parameters:
	specName string,
	workType WorkType,
	urlOrContents string,
	contentsSha256 string,
	metadataUrl string,
	mutable bool,
	// Accounts:
	groupAuthority ag_solanago.PublicKey,
	spec ag_solanago.PublicKey,
	workGroup ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CreateWorkSpec {
	return NewCreateWorkSpecInstructionBuilder().
		SetSpecName(specName).
		SetWorkType(workType).
		SetUrlOrContents(urlOrContents).
		SetContentsSha256(contentsSha256).
		SetMetadataUrl(metadataUrl).
		SetMutable(mutable).
		SetGroupAuthorityAccount(groupAuthority).
		SetSpecAccount(spec).
		SetWorkGroupAccount(workGroup).
		SetSystemProgramAccount(systemProgram)
}
