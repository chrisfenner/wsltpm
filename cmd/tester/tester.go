package main

import (
	"fmt"
	"os"

	"github.com/chrisfenner/wsltpm/pkg/tbs"
)

func mainErr() (err error) {
	err = nil
	ctx, err := tbs.Open()
	if err != nil {
		return
	}
	defer func() {
		closeErr := ctx.Close()
		// Only return the error from closing if there wasn't another one.
		if err == nil {
			err = closeErr
		}
	}()

	info, err := ctx.GetDeviceInfo()
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", *info)

	// PCR_Read
	cmd := []byte{
		0x80, 0x01,
		0x00, 0x00, 0x00, 0x14,
		0x00, 0x00, 0x01, 0x7E,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x0B,
		0x03,
		0xFF, 0xFF, 0xFF,
	}
	fmt.Printf("Submitting PCR_Read command...\n")
	rsp, err := ctx.SubmitCommand(cmd)
	if err != nil {
		return
	}
	fmt.Printf("PCR_Read response: %x\n", rsp)
	return
}

func main() {
	err := mainErr()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
