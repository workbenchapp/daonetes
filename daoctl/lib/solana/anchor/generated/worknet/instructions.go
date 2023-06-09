// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package worknet

import (
	"bytes"
	"fmt"
	ag_spew "github.com/davecgh/go-spew/spew"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_text "github.com/gagliardetto/solana-go/text"
	ag_treeout "github.com/gagliardetto/treeout"
)

var ProgramID ag_solanago.PublicKey

func SetProgramID(pubkey ag_solanago.PublicKey) {
	ProgramID = pubkey
	ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
}

const ProgramName = "Worknet"

func init() {
	if !ProgramID.IsZero() {
		ag_solanago.RegisterInstructionDecoder(ProgramID, registryDecodeInstruction)
	}
}

var (
	Instruction_CreateWorkGroup = ag_binary.TypeID([8]byte{80, 126, 215, 230, 103, 167, 176, 145})

	Instruction_CloseWorkGroup = ag_binary.TypeID([8]byte{90, 246, 55, 180, 72, 115, 56, 231})

	Instruction_RegisterDevice = ag_binary.TypeID([8]byte{210, 151, 56, 68, 22, 158, 90, 193})

	Instruction_CloseDevice = ag_binary.TypeID([8]byte{156, 69, 71, 242, 206, 207, 38, 134})

	Instruction_UpdateDevice = ag_binary.TypeID([8]byte{30, 154, 166, 184, 200, 216, 194, 74})

	Instruction_CreateWorkSpec = ag_binary.TypeID([8]byte{119, 183, 41, 136, 102, 42, 255, 215})

	Instruction_CloseWorkSpec = ag_binary.TypeID([8]byte{68, 1, 193, 232, 31, 12, 231, 113})

	Instruction_CreateDeployment = ag_binary.TypeID([8]byte{55, 207, 186, 101, 21, 218, 102, 171})

	Instruction_CloseDeployment = ag_binary.TypeID([8]byte{0, 65, 162, 218, 47, 208, 26, 62})

	Instruction_Schedule = ag_binary.TypeID([8]byte{149, 203, 229, 209, 47, 51, 221, 206})
)

// InstructionIDToName returns the name of the instruction given its ID.
func InstructionIDToName(id ag_binary.TypeID) string {
	switch id {
	case Instruction_CreateWorkGroup:
		return "CreateWorkGroup"
	case Instruction_CloseWorkGroup:
		return "CloseWorkGroup"
	case Instruction_RegisterDevice:
		return "RegisterDevice"
	case Instruction_CloseDevice:
		return "CloseDevice"
	case Instruction_UpdateDevice:
		return "UpdateDevice"
	case Instruction_CreateWorkSpec:
		return "CreateWorkSpec"
	case Instruction_CloseWorkSpec:
		return "CloseWorkSpec"
	case Instruction_CreateDeployment:
		return "CreateDeployment"
	case Instruction_CloseDeployment:
		return "CloseDeployment"
	case Instruction_Schedule:
		return "Schedule"
	default:
		return ""
	}
}

type Instruction struct {
	ag_binary.BaseVariant
}

func (inst *Instruction) EncodeToTree(parent ag_treeout.Branches) {
	if enToTree, ok := inst.Impl.(ag_text.EncodableToTree); ok {
		enToTree.EncodeToTree(parent)
	} else {
		parent.Child(ag_spew.Sdump(inst))
	}
}

var InstructionImplDef = ag_binary.NewVariantDefinition(
	ag_binary.AnchorTypeIDEncoding,
	[]ag_binary.VariantType{
		{
			"create_work_group", (*CreateWorkGroup)(nil),
		},
		{
			"close_work_group", (*CloseWorkGroup)(nil),
		},
		{
			"register_device", (*RegisterDevice)(nil),
		},
		{
			"close_device", (*CloseDevice)(nil),
		},
		{
			"update_device", (*UpdateDevice)(nil),
		},
		{
			"create_work_spec", (*CreateWorkSpec)(nil),
		},
		{
			"close_work_spec", (*CloseWorkSpec)(nil),
		},
		{
			"create_deployment", (*CreateDeployment)(nil),
		},
		{
			"close_deployment", (*CloseDeployment)(nil),
		},
		{
			"schedule", (*Schedule)(nil),
		},
	},
)

func (inst *Instruction) ProgramID() ag_solanago.PublicKey {
	return ProgramID
}

func (inst *Instruction) Accounts() (out []*ag_solanago.AccountMeta) {
	return inst.Impl.(ag_solanago.AccountsGettable).GetAccounts()
}

func (inst *Instruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := ag_binary.NewBorshEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *Instruction) TextEncode(encoder *ag_text.Encoder, option *ag_text.Option) error {
	return encoder.Encode(inst.Impl, option)
}

func (inst *Instruction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) error {
	return inst.BaseVariant.UnmarshalBinaryVariant(decoder, InstructionImplDef)
}

func (inst *Instruction) MarshalWithEncoder(encoder *ag_binary.Encoder) error {
	err := encoder.WriteBytes(inst.TypeID.Bytes(), false)
	if err != nil {
		return fmt.Errorf("unable to write variant type: %w", err)
	}
	return encoder.Encode(inst.Impl)
}

func registryDecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (interface{}, error) {
	inst, err := DecodeInstruction(accounts, data)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func DecodeInstruction(accounts []*ag_solanago.AccountMeta, data []byte) (*Instruction, error) {
	inst := new(Instruction)
	if err := ag_binary.NewBorshDecoder(data).Decode(inst); err != nil {
		return nil, fmt.Errorf("unable to decode instruction: %w", err)
	}
	if v, ok := inst.Impl.(ag_solanago.AccountsSettable); ok {
		err := v.SetAccounts(accounts)
		if err != nil {
			return nil, fmt.Errorf("unable to set accounts for instruction: %w", err)
		}
	}
	return inst, nil
}
