//go:build windows

package wmi

import (
	"testing"
)

func TestConnect(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	s.Close()
}

func TestQuery(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	result, err := s.Query("SELECT * FROM Win32_BIOS")
	if err != nil {
		t.Fatalf("Query() failed: %v", err)
	}
	if len(result.Columns) == 0 {
		t.Errorf("Query() returned no columns")
	}
	if len(result.Rows) == 0 {
		t.Errorf("Query() returned no rows")
	}
	for _, row := range result.Rows {
		t.Logf("%+v", row)
	}
}

func TestQueryStruct(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	var bios []BIOS
	err = s.QueryStruct("SELECT * FROM Win32_BIOS", &bios)
	if err != nil {
		t.Fatalf("QueryStruct() failed: %v", err)
	}
	if len(bios) == 0 {
		t.Errorf("QueryStruct() returned no results")
	}
	for _, b := range bios {
		t.Logf("BIOS: Manufacturer=%s Version=%s Serial=%s", b.Manufacturer, b.Version, b.SerialNumber)
	}
}

func TestQueryBIOS(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	bios, err := s.QueryBIOS()
	if err != nil {
		t.Fatalf("QueryBIOS() failed: %v", err)
	}
	if len(bios) == 0 {
		t.Errorf("QueryBIOS() returned no results")
	}
	for _, b := range bios {
		if b.Manufacturer == "" {
			t.Errorf("BIOS Manufacturer is empty")
		}
		if b.Version == "" {
			t.Errorf("BIOS Version is empty")
		}
		t.Logf("BIOS: Manufacturer=%s Name=%s Version=%s Serial=%s SMBIOSVersion=%s",
			b.Manufacturer, b.Name, b.Version, b.SerialNumber, b.SMBIOSBIOSVersion)
	}
}

func TestQueryComputerSystem(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	cs, err := s.QueryComputerSystem()
	if err != nil {
		t.Fatalf("QueryComputerSystem() failed: %v", err)
	}
	if len(cs) == 0 {
		t.Errorf("QueryComputerSystem() returned no results")
	}
	for _, c := range cs {
		if c.Manufacturer == "" {
			t.Errorf("ComputerSystem Manufacturer is empty")
		}
		if c.Model == "" {
			t.Errorf("ComputerSystem Model is empty")
		}
		t.Logf("ComputerSystem: Manufacturer=%s Model=%s SystemType=%s Domain=%s Processors=%d Logical=%d Memory=%d",
			c.Manufacturer, c.Model, c.SystemType, c.Domain, c.NumberOfProcessors, c.NumberOfLogicalProcessors, c.TotalPhysicalMemory)
	}
}

func TestQueryProcessor(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	procs, err := s.QueryProcessor()
	if err != nil {
		t.Fatalf("QueryProcessor() failed: %v", err)
	}
	if len(procs) == 0 {
		t.Errorf("QueryProcessor() returned no results")
	}
	for _, p := range procs {
		if p.Name == "" {
			t.Errorf("Processor Name is empty")
		}
		if p.Manufacturer == "" {
			t.Errorf("Processor Manufacturer is empty")
		}
		t.Logf("Processor: Name=%s Manufacturer=%s Cores=%d Threads=%d MaxClock=%d MHz Arch=%d",
			p.Name, p.Manufacturer, p.CoreCount, p.ThreadCount, p.MaxClockSpeed, p.Architecture)
	}
}

func TestQueryPhysicalMemory(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	mem, err := s.QueryPhysicalMemory()
	if err != nil {
		t.Fatalf("QueryPhysicalMemory() failed: %v", err)
	}
	t.Logf("PhysicalMemory: found %d module(s)", len(mem))
	for _, m := range mem {
		t.Logf("  Bank=%s Capacity=%d Speed=%d MHz Manufacturer=%s Part=%s Serial=%s",
			m.BankLabel, m.Capacity, m.Speed, m.Manufacturer, m.PartNumber, m.SerialNumber)
	}
}

func TestQueryDiskDrive(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	disks, err := s.QueryDiskDrive()
	if err != nil {
		t.Fatalf("QueryDiskDrive() failed: %v", err)
	}
	t.Logf("DiskDrive: found %d drive(s)", len(disks))
	for _, d := range disks {
		t.Logf("  Model=%s Interface=%s MediaType=%s Size=%d Partitions=%d Serial=%s",
			d.Model, d.InterfaceType, d.MediaType, d.Size, d.Partitions, d.SerialNumber)
	}
}

func TestQueryLogicalDisk(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	disks, err := s.QueryLogicalDisk()
	if err != nil {
		t.Fatalf("QueryLogicalDisk() failed: %v", err)
	}
	if len(disks) == 0 {
		t.Errorf("QueryLogicalDisk() returned no results")
	}
	foundC := false
	for _, d := range disks {
		if d.DeviceID == "C:" {
			foundC = true
		}
		t.Logf("  Device=%s FileSystem=%s Size=%d Free=%d Volume=%s Type=%d",
			d.DeviceID, d.FileSystem, d.Size, d.FreeSpace, d.VolumeName, d.DriveType)
	}
	if !foundC {
		t.Errorf("QueryLogicalDisk() did not find C: drive")
	}
}

func TestQueryNetworkAdapter(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	adapters, err := s.QueryNetworkAdapter()
	if err != nil {
		t.Fatalf("QueryNetworkAdapter() failed: %v", err)
	}
	t.Logf("NetworkAdapter: found %d adapter(s)", len(adapters))
	for _, a := range adapters {
		t.Logf("  Name=%s MAC=%s Type=%s Speed=%d Enabled=%t Manufacturer=%s Product=%s",
			a.Name, a.MACAddress, a.AdapterType, a.Speed, a.NetEnabled, a.Manufacturer, a.ProductName)
	}
}

func TestQueryOperatingSystem(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	osInfo, err := s.QueryOperatingSystem()
	if err != nil {
		t.Fatalf("QueryOperatingSystem() failed: %v", err)
	}
	if len(osInfo) == 0 {
		t.Errorf("QueryOperatingSystem() returned no results")
	}
	for _, o := range osInfo {
		if o.Caption == "" {
			t.Errorf("OperatingSystem Caption is empty")
		}
		if o.Version == "" {
			t.Errorf("OperatingSystem Version is empty")
		}
		if o.BuildNumber == "" {
			t.Errorf("OperatingSystem BuildNumber is empty")
		}
		if o.OSArchitecture == "" {
			t.Errorf("OperatingSystem OSArchitecture is empty")
		}
		t.Logf("OperatingSystem: Caption=%s Version=%s Build=%s Arch=%s Registered=%s",
			o.Caption, o.Version, o.BuildNumber, o.OSArchitecture, o.RegisteredUser)
	}
}

func TestQueryVideoController(t *testing.T) {
	s, err := Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer s.Close()

	vcs, err := s.QueryVideoController()
	if err != nil {
		t.Fatalf("QueryVideoController() failed: %v", err)
	}
	t.Logf("VideoController: found %d adapter(s)", len(vcs))
	for _, v := range vcs {
		t.Logf("  Name=%s RAM=%d Driver=%s Processor=%s Mode=%s Refresh=%d Hz",
			v.Name, v.AdapterRAM, v.DriverVersion, v.VideoProcessor, v.VideoModeDescription, v.CurrentRefreshRate)
	}
}
