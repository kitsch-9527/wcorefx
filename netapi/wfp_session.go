//go:build windows

package netapi

import (
	"fmt"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WfpSession 封装 WFP 引擎会话生命周期，避免重复打开/关闭引擎。
type WfpSession struct {
	engine windows.Handle
	closed bool
}

// NewWfpSession 创建并打开一个 WFP 引擎会话。
func NewWfpSession() (*WfpSession, error) {
	session := FwpmSession0{
		DisplayData: FwpmDisplayData0{
			Name: windows.StringToUTF16Ptr("wcorefx"),
		},
	}
	h, err := FwpmEngineOpen(nil, 10, nil, &session)
	if err != nil {
		return nil, fmt.Errorf("FwpmEngineOpen: %w", err)
	}
	return &WfpSession{engine: h}, nil
}

// Close 关闭 WFP 引擎会话。
func (s *WfpSession) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	return FwpmEngineClose(s.engine)
}

// Callouts 枚举所有 WFP 标注。
func (s *WfpSession) Callouts() ([]FwpmCallout, error) {
	enumH, err := FwpmCalloutCreateEnumHandle(s.engine, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutCreateEnumHandle: %w", err)
	}
	defer FwpmCalloutDestroyEnumHandle(s.engine, enumH)

	var array **FwpmCallout0
	var numEntries uint32
	err = FwpmCalloutEnum(s.engine, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutEnum: %w", err)
	}
	if numEntries == 0 {
		return nil, nil
	}

	entries := sliceFromArray(array, numEntries)
	defer FwpmFreeMemory(unsafe.Pointer(array))

	result := make([]FwpmCallout, 0, numEntries)
	for _, c := range entries {
		result = append(result, FwpmCallout{
			CalloutKey:   c.CalloutKey,
			CalloutId:    c.CalloutId,
			Name:         windows.UTF16PtrToString(c.DisplayData.Name),
			Description:  windows.UTF16PtrToString(c.DisplayData.Description),
			Flags:        c.Flags,
			ProviderKey:  c.ProviderKey,
			ProviderData: c.ProviderData,
			LayerKey:     c.LayerKey,
		})
	}
	return result, nil
}

// Filters 枚举所有 WFP 过滤器。
func (s *WfpSession) Filters() ([]FwpmFilter, error) {
	enumH, err := FwpmFilterCreateEnumHandle(s.engine, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterCreateEnumHandle: %w", err)
	}
	defer FwpmFilterDestroyEnumHandle(s.engine, enumH)

	var array **FwpmFilter0
	var numEntries uint32
	err = FwpmFilterEnum(s.engine, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterEnum: %w", err)
	}
	if numEntries == 0 {
		return nil, nil
	}

	entries := sliceFromArray(array, numEntries)
	defer FwpmFreeMemory(unsafe.Pointer(array))

	result := make([]FwpmFilter, 0, numEntries)
	for _, f := range entries {
		result = append(result, FwpmFilter{
			FilterKey:           f.FilterKey,
			Name:                windows.UTF16PtrToString(f.DisplayData.Name),
			Description:         windows.UTF16PtrToString(f.DisplayData.Description),
			Flags:               f.Flags,
			ProviderKey:         f.ProviderKey,
			ProviderData:        f.ProviderData,
			LayerKey:            f.LayerKey,
			SublayerKey:         f.SublayerKey,
			Weight:              f.Weight,
			NumFilterConditions: f.NumFilterConditions,
			FilterConditions:    f.FilterConditions,
			Action:              f.Action,
			RawContext:          f.RawContext,
			ProviderContextKey:  f.ProviderContextKey,
			Reserved:            f.Reserved,
			FilterID:            f.FilterID,
			EffectiveWeight:     f.EffectiveWeight,
		})
	}
	return result, nil
}

// sliceFromArray converts a C-style array pointer to a Go slice.
func sliceFromArray[T any](array **T, count uint32) []*T {
	if count == 0 {
		return nil
	}
	s := []*T{}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Cap = int(count)
	sh.Len = int(count)
	sh.Data = uintptr(unsafe.Pointer(array))
	return s
}
