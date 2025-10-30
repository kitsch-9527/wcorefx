//go:build windows
// +build windows

// Copyright (c) 2021 The Inet.Af AUTHORS. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package fwpuclnt

import (
	"fmt"

	"golang.org/x/sys/windows"
)

//go:notinheap
type FwpmDisplayData0 struct {
	Name        *uint16
	Description *uint16
}

type FwpmSession0Flags uint32

const FwpmSession0FlagDynamic = 1

//go:notinheap
type FwpmSession0 struct {
	SessionKey           windows.GUID
	DisplayData          FwpmDisplayData0
	Flags                FwpmSession0Flags
	TxnWaitTimeoutMillis uint32
	ProcessID            uint32
	SID                  *windows.SID
	Username             *uint16
	KernelMode           uint8
}

type authnService uint32

const (
	authnServiceWinNT   authnService = 0xa
	authnServiceDefault authnService = 0xffffffff
)

//go:notinheap
type fwpmLayerEnumTemplate0 struct {
	reserved uint64
}

//go:notinheap
type fwpmLayer0 struct {
	LayerKey           LayerID
	DisplayData        FwpmDisplayData0
	Flags              uint32
	NumFields          uint32
	Fields             *fwpmField0
	DefaultSublayerKey SublayerID
	LayerID            uint16
}

type fwpmFieldType uint32

const (
	fwpmFieldTypeRawData   fwpmFieldType = iota // no special semantics
	fwpmFieldTypeIPAddress                      // data is an IP address
	fwpmFieldTypeFlags                          // data is a flag bitfield
)

type dataType uint32

const (
	dataTypeEmpty                  dataType = 0
	dataTypeUint8                  dataType = 1
	dataTypeUint16                 dataType = 2
	dataTypeUint32                 dataType = 3
	dataTypeUint64                 dataType = 4
	dataTypeByteArray16            dataType = 11
	dataTypeByteBlob               dataType = 12
	dataTypeSID                    dataType = 13
	dataTypeSecurityDescriptor     dataType = 14
	dataTypeTokenInformation       dataType = 15
	dataTypeTokenAccessInformation dataType = 16
	dataTypeArray6                 dataType = 18
	dataTypeBitmapIndex            dataType = 19
	dataTypeV4AddrMask             dataType = 256
	dataTypeV6AddrMask             dataType = 257
	dataTypeRange                  dataType = 258
)

// Types not implemented, because WFP doesn't seem to use them.
// dataTypeInt8 dataType = 5
// dataTypeInt16 dataType = 6
// dataTypeInt32 dataType = 7
// dataTypeInt64 dataType = 8
// dataTypeFloat dataType = 9
// dataTypeDouble dataType = 10
// dataTypeUnicodeString dataType = 17
// dataTypeBitmapArray64 dataType = 20

//go:notinheap
type fwpmField0 struct {
	FieldKey *FieldID
	Type     fwpmFieldType
	DataType dataType
}

//go:notinheap
type fwpmSublayerEnumTemplate0 struct {
	ProviderKey *windows.GUID
}

//go:notinheap
type FwpByteBlob struct {
	Size uint32
	Data *uint8
}

type fwpmSublayerFlags uint32

const fwpmSublayerFlagsPersistent fwpmSublayerFlags = 1

//go:notinheap
type fwpmSublayer0 struct {
	SublayerKey  SublayerID
	DisplayData  FwpmDisplayData0
	Flags        fwpmSublayerFlags
	ProviderKey  *windows.GUID
	ProviderData FwpByteBlob
	Weight       uint16
}

type fwpmProviderFlags uint32

const (
	fwpmProviderFlagsPersistent fwpmProviderFlags = 0x01
	fwpmProviderFlagsDisabled   fwpmProviderFlags = 0x10
)

// ProviderID identifies a WFP provider.
type ProviderID windows.GUID

//go:notinheap
type fwpmProvider0 struct {
	ProviderKey  ProviderID
	DisplayData  FwpmDisplayData0
	Flags        fwpmProviderFlags
	ProviderData FwpByteBlob
	ServiceName  *uint16
}

//go:notinheap
type FwpValue0 struct {
	Type  dataType
	Value uintptr // unioned value
}

type FwpmFilterFlags uint32

const (
	FwpmFilterFlagsPersistent FwpmFilterFlags = 1 << iota
	FwpmFilterFlagsBootTime
	FwpmFilterFlagsHasProviderContext
	FwpmFilterFlagsClearActionRight
	FwpmFilterFlagsPermitIfCalloutUnregistered
	FwpmFilterFlagsDisabled
	FwpmFilterFlagsIndexed
)

// SublayerID identifies a WFP sublayer.
type SublayerID windows.GUID

// Action is an action the filtering engine can execute.
type Action uint32

const (
	// ActionBlock blocks a packet or session.
	ActionBlock Action = 0x1001
	// ActionPermit permits a packet or session.
	ActionPermit Action = 0x1002
	// ActionCalloutTerminating invokes a callout that must return a
	// permit or block verdict.
	ActionCalloutTerminating Action = 0x5003
	// ActionCalloutInspection invokes a callout that is expected to
	// not return a verdict (i.e. a read-only callout).
	ActionCalloutInspection Action = 0x6004
	// ActionCalloutUnknown invokes a callout that may return a permit
	// or block verdict.
	ActionCalloutUnknown Action = 0x4005
)

//go:notinheap
type FwpmAction0 struct {
	Type Action
	GUID windows.GUID
}
type RuleID windows.GUID

// FwpmFilter0 is the Go representation of FWPM_FILTER0,
// which stores the state associated with a filter.
// See https://docs.microsoft.com/en-us/windows/win32/api/fwpmtypes/ns-fwpmtypes-fwpm_filter0
//
//go:notinheap
type FwpmFilter0 struct {
	FilterKey           windows.GUID
	DisplayData         FwpmDisplayData0
	Flags               FwpmFilterFlags
	ProviderKey         *windows.GUID
	ProviderData        FwpByteBlob
	LayerKey            LayerID
	SublayerKey         SublayerID
	Weight              FwpValue0
	NumFilterConditions uint32
	FilterConditions    *FwpmFilterCondition0
	Action              FwpmAction0

	// Only one of RawContext/ProviderContextKey must be set.
	RawContext         uint64
	ProviderContextKey windows.GUID

	Reserved        *windows.GUID
	FilterID        uint64
	EffectiveWeight FwpValue0
}
type FwpmCallout0 struct {
	CalloutKey   windows.GUID
	DisplayData  FwpmDisplayData0
	Flags        FwpmFilterFlags
	ProviderKey  *windows.GUID
	ProviderData FwpByteBlob
	LayerKey     LayerID
	CalloutId    uint32
}

// LayerID identifies a WFP layer.
type LayerID windows.GUID

//go:notinheap
type fwpConditionValue0 struct {
	Type  dataType
	Value uintptr
}

// MatchType is the operator to use when testing a field in a Match.
type MatchType uint32 // do not change type, used in C calls

const (
	MatchTypeEqual MatchType = iota
	MatchTypeGreater
	MatchTypeLess
	MatchTypeGreaterOrEqual
	MatchTypeLessOrEqual
	MatchTypeRange // true if the field value is within the Range.
	MatchTypeFlagsAllSet
	MatchTypeFlagsAnySet
	MatchTypeFlagsNoneSet
	MatchTypeEqualCaseInsensitive // only valid on strings, no string fields exist
	MatchTypeNotEqual
	MatchTypePrefix    // TODO: not well documented. Is this prefix.Contains(ip) ?
	MatchTypeNotPrefix // TODO: see above.
)

var mtStr = map[MatchType]string{
	MatchTypeEqual:                "==",
	MatchTypeGreater:              ">",
	MatchTypeLess:                 "<",
	MatchTypeGreaterOrEqual:       ">=",
	MatchTypeLessOrEqual:          "<=",
	MatchTypeRange:                "in",
	MatchTypeFlagsAllSet:          "F[all]",
	MatchTypeFlagsAnySet:          "F[any]",
	MatchTypeFlagsNoneSet:         "F[none]",
	MatchTypeEqualCaseInsensitive: "i==",
	MatchTypeNotEqual:             "!=",
	MatchTypePrefix:               "pfx",
	MatchTypeNotPrefix:            "!pfx",
}

func (m MatchType) String() string {
	return mtStr[m]
}

// Match is a matching test that gets run against a layer's field.
type Match struct {
	Field FieldID
	Op    MatchType
	Value interface{}
}

func (m Match) String() string {
	return fmt.Sprintf("%s %s %v (%T)", m.Field, m.Op, m.Value, m.Value)
}

// FieldID identifies a WFP layer field.
type FieldID windows.GUID

//go:notinheap
type FwpmFilterCondition0 struct {
	FieldKey  FieldID
	MatchType MatchType
	Value     fwpConditionValue0
}

//go:notinheap
type fwpV4AddrAndMask struct {
	Addr, Mask uint32
}

//go:notinheap
type fwpV6AddrAndMask struct {
	Addr         [16]byte
	PrefixLength uint8
}

//go:notinheap
type fwpmProviderContextEnumTemplate0 struct {
	ProviderKey         *ProviderID
	ProviderContextType uint32
}

//go:notinheap
type fwpmFilterEnumTemplate0 struct {
	ProviderKey             *ProviderID
	LayerKey                windows.GUID
	EnumType                filterEnumType
	Flags                   filterEnumFlags
	ProviderContextTemplate *fwpmProviderContextEnumTemplate0 // TODO: wtf?
	NumConditions           uint32
	Conditions              *FwpmFilterCondition0
	ActionMask              uint32
	CalloutKey              *windows.GUID
}

//go:notinheap
type fwpRange0 struct {
	From, To FwpValue0
}

type filterEnumType uint32

const (
	filterEnumTypeFullyContained filterEnumType = iota
	filterEnumTypeOverlapping
)

type filterEnumFlags uint32

const (
	filterEnumFlagsBestTerminatingMatch filterEnumFlags = iota + 1
	filterEnumFlagsSorted
	filterEnumFlagsBootTimeOnly
	filterEnumFlagsIncludeBootTime
	filterEnumFlagsIncludeDisabled
)

type fwpIPVersion uint32

const (
	fwpIPVersion4 fwpIPVersion = 0
	fwpIPVersion6 fwpIPVersion = 1
)

//go:notinheap
type fwpmNetEventHeader1 struct {
	Timestamp  windows.Filetime
	Flags      uint32       // enum
	IPVersion  fwpIPVersion // enum
	IPProtocol uint8
	_          [3]byte
	LocalAddr  [16]byte
	RemoteAddr [16]byte
	LocalPort  uint16
	RemotePort uint16
	ScopeID    uint32
	AppID      FwpByteBlob
	UserID     *windows.SID

	// Random reserved fields for an aborted attempt at including
	// Ethernet frame information. Not used, but we have to pad out
	// the struct appropriately.
	_ struct {
		reserved1 uint32
		unused2   struct {
			reserved2  [6]byte
			reserved3  [6]byte
			reserved4  uint32
			reserved5  uint32
			reserved6  uint16
			reserved7  uint32
			reserved8  uint32
			reserved9  uint16
			reserved10 uint64
		}
	}
}

//go:notinheap
type fwpmNetEventClassifyDrop1 struct {
	FilterID        uint64
	LayerID         uint16
	ReauthReason    uint32
	OriginalProfile uint32
	CurrentProfile  uint32
	Direction       uint32
	Loopback        uint32
}

type fwpmNetEventType uint32

const fwpmNetEventClassifyDrop = 3

//go:notinheap
type fwpmNetEvent1 struct {
	Header fwpmNetEventHeader1
	Type   fwpmNetEventType
	Drop   *fwpmNetEventClassifyDrop1
}

// // SEC_WINNT_AUTH_IDENTITY 定义Windows认证身份信息
// 参考: https://learn.microsoft.com/en-us/windows/win32/api/sspi/ns-sspi-sec_winnt_auth_identity_a
type secWinntAuthIdentity struct {
	User           *uint8
	UserLength     uint32
	Domain         *uint8
	DomainLength   uint32
	Password       *uint8
	PasswordLength uint32
	Flags          FwpmFilterFlags
}

// FWPM_CALLOUT_ENUM_TEMPLATE0 定义标注枚举模板
// 参考: https://learn.microsoft.com/en-us/windows/win32/api/fwpmtypes/ns-fwpmtypes-fwpm_callout_enum_template0
type fwpmCalloutEnumTemplate0 struct {
	ProviderKey *windows.GUID // 对应C中的GUID* providerKey，指向提供程序的GUID
	LayerKey    windows.GUID  // 对应C中的GUID layerKey，标注所在层的GUID
}
