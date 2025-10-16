// SPDX-License-Identifier: BSD-3-Clause
//go:build (windows && amd64) || (windows && arm64)

package ps

type PROCESS_MEMORY_COUNTERS struct { //nolint:revive //FIXME
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uint64
	WorkingSetSize             uint64
	QuotaPeakPagedPoolUsage    uint64
	QuotaPagedPoolUsage        uint64
	QuotaPeakNonPagedPoolUsage uint64
	QuotaNonPagedPoolUsage     uint64
	PagefileUsage              uint64
	PeakPagefileUsage          uint64
}
