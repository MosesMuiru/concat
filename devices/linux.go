package devices

import (
	"bufio"
	"os"
	"strings"

	"github.com/mosesmuiru/concat/devices/drive"
)

// --- Linux: parse /proc/mounts ---
func GetLinuxDrives() ([]drive.Drive, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var drives []drive.Drive

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		device, mountPoint, fsType := fields[0], fields[1], fields[2]

		// USB drives typically mount under /media or /mnt, on /dev/sd* or /dev/nvme*
		isExternal := (strings.HasPrefix(device, "/dev/sd") ||
			strings.HasPrefix(device, "/dev/nvme")) &&
			(strings.HasPrefix(mountPoint, "/media") ||
				strings.HasPrefix(mountPoint, "/mnt") ||
				strings.HasPrefix(mountPoint, "/run/media"))

		if isExternal {
			drives = append(drives, drive.Drive{device, mountPoint, fsType})
		}
	}
	return drives, scanner.Err()
}
