package eve

// EVT_EXPORTLOG_FLAGS 表示事件日志导出的标志类型
type EVT_EXPORTLOG_FLAGS uint32

const (
	// EvtExportLogChannelPath 表示导出通道路径
	EvtExportLogChannelPath EVT_EXPORTLOG_FLAGS = 0x1
	// EvtExportLogFilePath 表示导出文件路径
	EvtExportLogFilePath EVT_EXPORTLOG_FLAGS = 0x2
	// EvtExportLogTolerateQueryErrors 表示容忍查询错误
	EvtExportLogTolerateQueryErrors EVT_EXPORTLOG_FLAGS = 0x1000
	// EvtExportLogOverwrite 表示覆盖已有文件
	EvtExportLogOverwrite EVT_EXPORTLOG_FLAGS = 0x2000
)
