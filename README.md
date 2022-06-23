# WSL TPM

## Overview

This project allows Linux programs running in the
[Windows Subsytem for Linux](https://docs.microsoft.com/en-us/windows/wsl/about)
to access the host TPM.

WSL provides the capability to run Windows binaries from Linux. wsltpm is under
development and will eventually comprise two parts:

* cmd/wsltpm.exe (Windows binary) (WIP)
  * Proxies calls to
    [tbs.dll](https://docs.microsoft.com/en-us/windows/win32/api/_tbs/)
* pkg/wsltpm (Linux-native library) (WIP)
  * Implements [go-tpm](https://github.com/google/go-tpm)'s
    [transport.TPM](https://github.com/google/go-tpm/blob/3b95eba2a55716cb43587e7a2f1aa499240a1d9f/direct/transport/tpm.go#L10-L14)
    so that go-tpm code can use wsltpm.exe to talk to the host TPM.

## Building

Both the tester binary and the wsltpm binary need to be compiled for Windows.
This can be done from Linux with:

```
GOOS=windows go build ./cmd/tester
```

## Testing

Thanks to the magic of WSL, the Windows binary can be directly run, e.g.:

```
./tester.exe
```
