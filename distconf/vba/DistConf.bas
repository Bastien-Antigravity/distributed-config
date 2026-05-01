Attribute VB_Name = "DistConf"
' -----------------------------------------------------------------------------------------------
' Distributed Config VBA Binding
' -----------------------------------------------------------------------------------------------
' This module provides access to the libdistconf shared library from VBA (Excel, Access, etc.).
'
' SETUP:
' 1. Build the library using 'make build-lib'.
' 2. Ensure libdistconf.so (or .dylib) is in your library search path.
' -----------------------------------------------------------------------------------------------

Option Explicit

#If Mac Then
    Private Const LIB_PATH As String = "libdistconf.dylib"
#Else
    Private Const LIB_PATH As String = "libdistconf.dll"
#End If

' Bridge API Declarations (Standardized to Go Naming)
' -----------------------------------------------------------------------------------------------

Public Declare PtrSafe Function DistConf_New Lib LIB_PATH (ByVal profile As String) As LongPtr
Public Declare PtrSafe Sub DistConf_Close Lib LIB_PATH (ByVal handle As LongPtr)
Public Declare PtrSafe Function DistConf_Get Lib LIB_PATH (ByVal handle As LongPtr, ByVal section As String, ByVal key As String) As LongPtr
Public Declare PtrSafe Function DistConf_Set Lib LIB_PATH (ByVal handle As LongPtr, ByVal section As String, ByVal key As String, ByVal value As String) As Long
Public Declare PtrSafe Function DistConf_Sync Lib LIB_PATH (ByVal handle As LongPtr) As Long
Public Declare PtrSafe Function DistConf_ShareConfig Lib LIB_PATH (ByVal handle As LongPtr, ByVal jsonData As String) As Long
Public Declare PtrSafe Sub DistConf_OnLiveConfUpdate Lib LIB_PATH (ByVal handle As LongPtr, ByVal callbackAddr As LongPtr)
Public Declare PtrSafe Sub DistConf_OnRegistryUpdate Lib LIB_PATH (ByVal handle As LongPtr, ByVal callbackAddr As LongPtr)
Public Declare PtrSafe Function DistConf_ValidateMandatoryServices Lib LIB_PATH (ByVal handle As LongPtr) As Long
Public Declare PtrSafe Function DistConf_GetAddress Lib LIB_PATH (ByVal handle As LongPtr, ByVal capability As String) As LongPtr
Public Declare PtrSafe Function DistConf_GetGRPCAddress Lib LIB_PATH (ByVal handle As LongPtr, ByVal capability As String) As LongPtr
Public Declare PtrSafe Function DistConf_GetCapability Lib LIB_PATH (ByVal handle As LongPtr, ByVal capability As String) As LongPtr
Public Declare PtrSafe Function DistConf_Decrypt Lib LIB_PATH (ByVal handle As LongPtr, ByVal ciphertext As String) As LongPtr
Public Declare PtrSafe Sub DistConf_FreeString Lib LIB_PATH (ByVal ptr As LongPtr)

' -----------------------------------------------------------------------------------------------
' High-level API Example
' -----------------------------------------------------------------------------------------------

Public Sub TestDistConf()
    Dim handle As LongPtr
    handle = DistConf_New("standalone")
    
    If handle = 0 Then
        MsgBox "Failed to initialize DistConf"
        Exit Sub
    End If
    
    ' Set a value
    Dim result As Long
    result = DistConf_Set(handle, "vba_demo", "status", "active")
    
    ' Sync and Validate
    If DistConf_ValidateMandatoryServices(handle) = 1 Then
        Debug.Print "Environment Validated"
    End If
    
    ' Clean up
    DistConf_Close handle
End Sub
