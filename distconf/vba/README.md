# DistConf VBA SDK

Integrate Excel, Access, and other Office applications into the `distributed-config` ecosystem.

## Setup

1. Build `libdistconf.so` (or `.dylib` or `.dll`).
2. Import `DistConf.bas` into your VBA project.
3. On macOS, ensure the library is in a location accessible to Excel (usually needs a full path in the `Declare` statements).

## Usage

```vba
Sub Demo()
    Dim handle As LongPtr
    handle = DistConf_New("standalone")
    
    ' Get value
    MsgBox "Service: " & PtrToString(DistConf_Get(handle, "common", "name"))
    
    DistConf_Close handle
End Sub
```
