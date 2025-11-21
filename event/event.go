//go:build windows
// +build windows

package event

import (
	"fmt"
	"strings"

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

// RemoveWindowsLineEndings replaces carriage return line feed (CRLF) with
// line feed (LF) and trims any newline character that may exist at the end
// of the string.
func RemoveWindowsLineEndings(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	return strings.TrimRight(s, "\n")
}
