//go:build windows
// +build windows

package event

import (
	"fmt"

	"github.com/kitsch-9527/wcorefx/winapi/dll/wevtapi"
	"golang.org/x/sys/windows"
)

// ExportLogsToEvents 导出指定时间范围内的日志到文件
func ExportLogsToEvents(evType string, queryStr, outFileStr string) error {
	eveType := windows.StringToUTF16Ptr(string(evType))
	query := windows.StringToUTF16Ptr(queryStr)
	outFile := windows.StringToUTF16Ptr(outFileStr)
	err := wevtapi.EvtExportLog(0, eveType, query, outFile, wevtapi.EvtExportLogChannelPath)
	if err != nil {
		return fmt.Errorf("EvtExportLog failed with error: %v", err)
	}
	return nil
}
