//go:build windows
// +build windows

package eve

import (
	"fmt"

	win "golang.org/x/sys/windows"
)

// ExportLogsToEvents 导出指定时间范围内的日志到文件
func ExportLogsToEvents(evType string, queryStr, outFileStr string) error {
	eveType := win.StringToUTF16Ptr(string(evType))
	query := win.StringToUTF16Ptr(queryStr)
	outFile := win.StringToUTF16Ptr(outFileStr)
	err := EvtExportLog(0, eveType, query, outFile, EvtExportLogChannelPath)
	if err != nil {
		return fmt.Errorf("EvtExportLog failed with error: %v", err)
	}
	return nil
}

