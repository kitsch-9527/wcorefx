//go:build windows

package wmi

import "time"

// BIOS 表示 Win32_BIOS WMI 类信息
type BIOS struct {
	// Manufacturer 主板制造商名称
	Manufacturer string `wmi:"Manufacturer"`
	// Name BIOS 名称
	Name string `wmi:"Name"`
	// Version BIOS 版本号
	Version string `wmi:"Version"`
	// SerialNumber BIOS 序列号
	SerialNumber string `wmi:"SerialNumber"`
	// ReleaseDate BIOS 发布日期
	ReleaseDate time.Time `wmi:"ReleaseDate"`
	// SMBIOSBIOSVersion SMBIOS BIOS 版本字符串
	SMBIOSBIOSVersion string `wmi:"SMBIOSBIOSVersion"`
}

// ComputerSystem 表示 Win32_ComputerSystem WMI 类信息
type ComputerSystem struct {
	// Manufacturer 系统制造商名称
	Manufacturer string `wmi:"Manufacturer"`
	// Model 系统型号
	Model string `wmi:"Model"`
	// SystemType 系统类型（如 x64-based PC）
	SystemType string `wmi:"SystemType"`
	// TotalPhysicalMemory 物理内存总大小（字节）
	TotalPhysicalMemory uint64 `wmi:"TotalPhysicalMemory"`
	// NumberOfProcessors 物理处理器数量
	NumberOfProcessors uint32 `wmi:"NumberOfProcessors"`
	// NumberOfLogicalProcessors 逻辑处理器数量
	NumberOfLogicalProcessors uint32 `wmi:"NumberOfLogicalProcessors"`
	// Domain 计算机所属域
	Domain string `wmi:"Domain"`
	// PartOfDomain 是否属于域
	PartOfDomain bool `wmi:"PartOfDomain"`
}

// Processor 表示 Win32_Processor WMI 类信息
type Processor struct {
	// Name 处理器名称
	Name string `wmi:"Name"`
	// Manufacturer 处理器制造商
	Manufacturer string `wmi:"Manufacturer"`
	// MaxClockSpeed 处理器最大时钟频率（MHz）
	MaxClockSpeed uint32 `wmi:"MaxClockSpeed"`
	// CurrentClockSpeed 处理器当前时钟频率（MHz）
	CurrentClockSpeed uint32 `wmi:"CurrentClockSpeed"`
	// CoreCount 物理核心数量
	CoreCount uint32 `wmi:"NumberOfCores"`
	// ThreadCount 线程数量
	ThreadCount uint32 `wmi:"NumberOfLogicalProcessors"`
	// L2CacheSize L2 缓存大小（KB）
	L2CacheSize uint32 `wmi:"L2CacheSize"`
	// L3CacheSize L3 缓存大小（KB）
	L3CacheSize uint32 `wmi:"L3CacheSize"`
	// Architecture 处理器架构（0=x86, 1=MIPS, 2=Alpha, 3=PowerPC, 5=ARM, 6=ia64, 9=x64）
	Architecture uint16 `wmi:"Architecture"`
}

// PhysicalMemory 表示 Win32_PhysicalMemory WMI 类信息
type PhysicalMemory struct {
	// BankLabel 物理内存插槽标签
	BankLabel string `wmi:"BankLabel"`
	// Capacity 内存容量（字节）
	Capacity uint64 `wmi:"Capacity"`
	// Speed 内存速度（MHz）
	Speed uint32 `wmi:"Speed"`
	// MemoryType 内存类型（0=Unknown, 1=Other, 2=DRAM, 3=Synchronous, 4=Cache, 5=ECC, 6=EDO, 7=DDR, 8=DDR2, 9=DDR3, 10=DDR4）
	MemoryType uint32 `wmi:"MemoryType"`
	// Manufacturer 内存制造商
	Manufacturer string `wmi:"Manufacturer"`
	// PartNumber 内存部件号
	PartNumber string `wmi:"PartNumber"`
	// SerialNumber 内存序列号
	SerialNumber string `wmi:"SerialNumber"`
}

// DiskDrive 表示 Win32_DiskDrive WMI 类信息
type DiskDrive struct {
	// Model 磁盘驱动器型号
	Model string `wmi:"Model"`
	// InterfaceType 接口类型（如 IDE, SATA, SCSI）
	InterfaceType string `wmi:"InterfaceType"`
	// MediaType 介质类型（如 Fixed hard disk media）
	MediaType string `wmi:"MediaType"`
	// Size 磁盘总大小（字节）
	Size uint64 `wmi:"Size"`
	// Partitions 分区数量
	Partitions uint32 `wmi:"Partitions"`
	// BytesPerSector 每扇区字节数
	BytesPerSector uint32 `wmi:"BytesPerSector"`
	// SerialNumber 磁盘序列号
	SerialNumber string `wmi:"SerialNumber"`
}

// LogicalDisk 表示 Win32_LogicalDisk WMI 类信息
type LogicalDisk struct {
	// DeviceID 驱动器号（如 C:）
	DeviceID string `wmi:"DeviceID"`
	// FileSystem 文件系统类型（如 NTFS, FAT32）
	FileSystem string `wmi:"FileSystem"`
	// Size 逻辑磁盘总大小（字节）
	Size uint64 `wmi:"Size"`
	// FreeSpace 逻辑磁盘可用空间（字节）
	FreeSpace uint64 `wmi:"FreeSpace"`
	// VolumeName 卷标名称
	VolumeName string `wmi:"VolumeName"`
	// VolumeSerialNumber 卷序列号
	VolumeSerialNumber string `wmi:"VolumeSerialNumber"`
	// DriveType 驱动器类型（0=Unknown, 1=NoRootDir, 2=Removable, 3=Local, 4=Network, 5=CDROM, 6=RAMDisk）
	DriveType uint32 `wmi:"DriveType"`
}

// NetworkAdapter 表示 Win32_NetworkAdapter WMI 类信息
type NetworkAdapter struct {
	// Name 网络适配器名称
	Name string `wmi:"Name"`
	// MACAddress MAC 地址
	MACAddress string `wmi:"MACAddress"`
	// AdapterType 适配器类型（如 Ethernet 802.3）
	AdapterType string `wmi:"AdapterType"`
	// Speed 网络适配器速度（bits/秒）
	Speed uint64 `wmi:"Speed"`
	// NetEnabled 是否启用网络适配器
	NetEnabled bool `wmi:"NetEnabled"`
	// Manufacturer 网络适配器制造商
	Manufacturer string `wmi:"Manufacturer"`
	// ProductName 网络适配器产品名称
	ProductName string `wmi:"ProductName"`
}

// OperatingSystem 表示 Win32_OperatingSystem WMI 类信息
type OperatingSystem struct {
	// Caption 操作系统名称（如 Microsoft Windows 11 Pro）
	Caption string `wmi:"Caption"`
	// Version 操作系统版本号
	Version string `wmi:"Version"`
	// BuildNumber 操作系统构建号
	BuildNumber string `wmi:"BuildNumber"`
	// OSArchitecture 操作系统架构（如 64-bit）
	OSArchitecture string `wmi:"OSArchitecture"`
	// InstallDate 操作系统安装日期
	InstallDate time.Time `wmi:"InstallDate"`
	// LastBootUpTime 上次启动时间
	LastBootUpTime time.Time `wmi:"LastBootUpTime"`
	// RegisteredUser 注册用户
	RegisteredUser string `wmi:"RegisteredUser"`
	// SerialNumber 操作系统序列号
	SerialNumber string `wmi:"SerialNumber"`
}

// VideoController 表示 Win32_VideoController WMI 类信息
type VideoController struct {
	// Name 显卡名称
	Name string `wmi:"Name"`
	// AdapterRAM 显存大小（字节）
	AdapterRAM uint64 `wmi:"AdapterRAM"`
	// DriverVersion 显卡驱动版本
	DriverVersion string `wmi:"DriverVersion"`
	// VideoProcessor 显卡处理器描述
	VideoProcessor string `wmi:"VideoProcessor"`
	// VideoModeDescription 当前显示模式描述
	VideoModeDescription string `wmi:"VideoModeDescription"`
	// CurrentRefreshRate 当前刷新率（Hz）
	CurrentRefreshRate uint32 `wmi:"CurrentRefreshRate"`
}

// QueryBIOS 查询 Win32_BIOS 信息
func (s *Session) QueryBIOS() ([]BIOS, error) {
	var result []BIOS
	err := s.QueryStruct("SELECT * FROM Win32_BIOS", &result)
	return result, err
}

// QueryComputerSystem 查询 Win32_ComputerSystem 信息
func (s *Session) QueryComputerSystem() ([]ComputerSystem, error) {
	var result []ComputerSystem
	err := s.QueryStruct("SELECT * FROM Win32_ComputerSystem", &result)
	return result, err
}

// QueryProcessor 查询 Win32_Processor 信息
func (s *Session) QueryProcessor() ([]Processor, error) {
	var result []Processor
	err := s.QueryStruct("SELECT * FROM Win32_Processor", &result)
	return result, err
}

// QueryPhysicalMemory 查询 Win32_PhysicalMemory 信息
func (s *Session) QueryPhysicalMemory() ([]PhysicalMemory, error) {
	var result []PhysicalMemory
	err := s.QueryStruct("SELECT * FROM Win32_PhysicalMemory", &result)
	return result, err
}

// QueryDiskDrive 查询 Win32_DiskDrive 信息
func (s *Session) QueryDiskDrive() ([]DiskDrive, error) {
	var result []DiskDrive
	err := s.QueryStruct("SELECT * FROM Win32_DiskDrive", &result)
	return result, err
}

// QueryLogicalDisk 查询 Win32_LogicalDisk 信息
func (s *Session) QueryLogicalDisk() ([]LogicalDisk, error) {
	var result []LogicalDisk
	err := s.QueryStruct("SELECT * FROM Win32_LogicalDisk", &result)
	return result, err
}

// QueryNetworkAdapter 查询 Win32_NetworkAdapter 信息
func (s *Session) QueryNetworkAdapter() ([]NetworkAdapter, error) {
	var result []NetworkAdapter
	err := s.QueryStruct("SELECT * FROM Win32_NetworkAdapter", &result)
	return result, err
}

// QueryOperatingSystem 查询 Win32_OperatingSystem 信息
func (s *Session) QueryOperatingSystem() ([]OperatingSystem, error) {
	var result []OperatingSystem
	err := s.QueryStruct("SELECT * FROM Win32_OperatingSystem", &result)
	return result, err
}

// QueryVideoController 查询 Win32_VideoController 信息
func (s *Session) QueryVideoController() ([]VideoController, error) {
	var result []VideoController
	err := s.QueryStruct("SELECT * FROM Win32_VideoController", &result)
	return result, err
}
