//go:build windows

package netapi

import (
	"testing"
)

func TestTcp4Endpoints(t *testing.T) {
	rows, err := Tcp4Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("IPv4 TCP endpoints: %d", len(rows))
	for _, r := range rows {
		if r.DwOwningPid == 0 && r.DwState != MibTCPStateListen {
			continue
		}
		t.Logf("  PID=%d State=%s Local=%s:%d Remote=%s:%d",
			r.DwOwningPid, TCPState(r.DwState),
			InetNtoa(r.DwLocalAddr), Ntohs(r.DwLocalPort),
			InetNtoa(r.DwRemoteAddr), Ntohs(r.DwRemotePort))
	}
}

func TestTcp6Endpoints(t *testing.T) {
	rows, err := Tcp6Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("IPv6 TCP endpoints: %d", len(rows))
	for _, r := range rows {
		t.Logf("  PID=%d State=%s Local=%s Scope=%d",
			r.DwOwningPid, TCPState(r.DwState),
			InetNtoa6(r.LocalAddr), r.LocalScopeId)
	}
}

func TestUdp4Endpoints(t *testing.T) {
	rows, err := Udp4Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("IPv4 UDP endpoints: %d", len(rows))
	for _, r := range rows {
		t.Logf("  PID=%d Local=%s:%d",
			r.DwOwningPid, InetNtoa(r.DwLocalAddr), Ntohs(r.DwLocalPort))
	}
}

func TestUdp6Endpoints(t *testing.T) {
	rows, err := Udp6Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("IPv6 UDP endpoints: %d", len(rows))
}

func TestTCPState(t *testing.T) {
	tests := []struct {
		state uint32
		want  string
	}{
		{MibTCPStateClosed, "CLOSED"},
		{MibTCPStateListen, "LISTENING"},
		{MibTCPStateEstablished, "ESTABLISHED"},
		{MibTCPStateTimeWait, "TIME_WAIT"},
		{99, "UNKNOWN(99)"},
	}
	for _, tt := range tests {
		got := TCPState(tt.state)
		if got != tt.want {
			t.Errorf("TCPState(%d) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestInetNtoa(t *testing.T) {
	// 127.0.0.1 in network byte order = 0x7f000001
	ip := InetNtoa(0x7f000001)
	if ip != "127.0.0.1" {
		t.Errorf("InetNtoa = %q, want 127.0.0.1", ip)
	}
	// 0.0.0.0
	ip = InetNtoa(0)
	if ip != "0.0.0.0" {
		t.Errorf("InetNtoa(0) = %q, want 0.0.0.0", ip)
	}
}

func TestNtohs(t *testing.T) {
	port := Ntohs(0x230a) // big-endian 0x0a23 = 2595
	if port != 2595 {
		t.Errorf("Ntohs = %d, want 2595", port)
	}
}

func TestInetNtoa6(t *testing.T) {
	var addr [16]byte
	addr[0] = 0xfe
	addr[1] = 0x80
	addr[15] = 0x01
	s := InetNtoa6(addr)
	if s == "" {
		t.Error("InetNtoa6 returned empty")
	}
	t.Logf("IPv6 addr: %s", s)
}

func TestInterfaces(t *testing.T) {
	ifaces, err := Interfaces()
	if err != nil {
		t.Fatalf("Interfaces() failed: %v", err)
	}
	if len(ifaces) == 0 {
		t.Fatal("Interfaces() returned empty slice")
	}
	t.Logf("Found %d network interfaces", len(ifaces))
	for _, iface := range ifaces {
		if iface.Name != "" {
			t.Logf("  %s: %s, IP=%s, MAC=%X", iface.Name, iface.Description, iface.IP, iface.MAC)
			break
		}
	}
}

func TestARP(t *testing.T) {
	entries, err := ARP()
	if err != nil {
		t.Fatalf("ARP() failed: %v", err)
	}
	t.Logf("Found %d ARP entries", len(entries))
}

func TestRoute(t *testing.T) {
	entries, err := Route()
	if err != nil {
		t.Fatalf("Route() failed: %v", err)
	}
	t.Logf("Found %d route entries", len(entries))
}
