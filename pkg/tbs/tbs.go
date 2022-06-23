package tbs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func wrapDLLError(r1, r2 uintptr, err error) error {
	var errno windows.Errno
	if !errors.As(err, &errno) {
		return err
	}
	// Only interpret errors in the TBS/TPM error spaces
	// This is because sometimes we get back the length of the TPM response
	// as an HRESULT when calling TBS from Go. (TODO: understand why)
	if (errno>>16) == 0x8028 || (errno>>16) == 0x8029 {
		return err
	}
	if r1 != 0 {
		return fmt.Errorf("HRESULT: 0x%0x", r1)
	}
	if r2 != 0 {
		return fmt.Errorf("HRESULT: 0x%0x", r2)
	}
	return nil
}

type tbsContextParams2Flags uint32

const (
	requestRaw tbsContextParams2Flags = 1 << iota
	includeTpm12
	includeTpm20
)

type tbsContextParams2 struct {
	version uint32
	flags   tbsContextParams2Flags
}

type context struct {
	dll *windows.DLL
	ctx unsafe.Pointer
}

// Open opens TBS and creates a context for calling TPM commands.
func Open() (*context, error) {
	tbs, err := windows.LoadDLL("tbs.dll")
	if err != nil {
		return nil, fmt.Errorf("error loading TBS.dll: %w", err)
	}
	proc, err := tbs.FindProc("Tbsi_Context_Create")
	if err != nil {
		return nil, fmt.Errorf("error finding Tbsi_Context_Create: %w", err)
	}
	tcp := tbsContextParams2{
		version: 2,
		flags:   includeTpm20,
	}
	var ctx unsafe.Pointer
	err = wrapDLLError(proc.Call(
		uintptr(unsafe.Pointer(&tcp)),
		uintptr(unsafe.Pointer(&ctx))))
	if err != nil {
		return nil, fmt.Errorf("error calling Tbsi_Context_Create: %w", err)
	}
	return &context{
		dll: tbs,
		ctx: ctx,
	}, nil
}

// Close releases the TBS context and handle to the DLL.
func (c *context) Close() error {
	proc, err := c.dll.FindProc("Tbsip_Context_Close")
	if err != nil {
		return fmt.Errorf("error finding Tbsip_Context_Close: %w")
	}
	err = wrapDLLError(proc.Call(uintptr(c.ctx)))
	if err != nil {
		return fmt.Errorf("error calling Tbsip_Context_Close: %w", err)
	}
	return c.dll.Release()
}

type DeviceInfo struct {
	StructVersion    uint32
	TpmVersion       uint32
	TpmInterfaceType uint32
	TpmImpRevision   uint32
}

// GetDeviceInfo calls Tbsi_GetDeviceInfo to fetch information about the TPM.
func (c *context) GetDeviceInfo() (*DeviceInfo, error) {
	proc, err := c.dll.FindProc("Tbsi_GetDeviceInfo")
	if err != nil {
		return nil, fmt.Errorf("error finding Tbsi_GetDeviceInfo: %w", err)
	}
	info := make([]byte, 16)
	err = wrapDLLError(proc.Call(16, uintptr(unsafe.Pointer(&info[0]))))
	if err != nil {
		return nil, fmt.Errorf("error calling Tbsi_GetDeviceInfo: %w", err)
	}
	var infoStruct DeviceInfo
	err = binary.Read(bytes.NewReader(info), binary.LittleEndian, &infoStruct)
	if err != nil {
		return nil, fmt.Errorf("error parsing result of Tbsi_GetDeviceInfo: %w", err)
	}
	return &infoStruct, nil
}

// SubmitCommand calls Tbsip_SubmitCommand to send the command to the TPM.
func (c *context) SubmitCommand(cmd []byte) ([]byte, error) {
	proc, err := c.dll.FindProc("Tbsip_Submit_Command")
	if err != nil {
		return nil, fmt.Errorf("error finding Tbsip_Submit_Command: %w", err)
	}
	rsp := make([]byte, 4096)
	rspLen := uint32(len(rsp))
	err = wrapDLLError(proc.Call(
		uintptr(c.ctx),
		0,
		200, // TBS_COMMAND_PRIORITY_NORMAL
		uintptr(unsafe.Pointer(&cmd[0])),
		uintptr(len(cmd)),
		uintptr(unsafe.Pointer(&rsp[0])),
		uintptr(unsafe.Pointer(&rspLen))))
	if err != nil {
		return nil, fmt.Errorf("error calling Tbsip_Submit_Command: %w", err)
	}
	rsp = rsp[:rspLen]
	return rsp, nil
}
