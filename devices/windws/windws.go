package windws

import (
	"fmt"

	"golang.org/x/sys/windows"

	"github.com/mosesmuiru/concat/devices/drive"
)


func GetWindowsDrives() ([]drive.Drive, error) {
    mask, err := windows.GetLogicalDrives()
    if err != nil {
        return nil, fmt.Errorf("GetLogicalDrives: %w", err)
    }

    var drives []drive.Drive

    for i := 0; i < 26; i++ {
        if mask&(1<<uint(i)) == 0 {
            continue
        }

        letter := string(rune('A' + i))
        rootPath := letter + `:\`

        rootPtr, _ := windows.UTF16PtrFromString(rootPath)

        driveType := windows.GetDriveType(rootPtr)
        if driveType != windows.DRIVE_REMOVABLE {
            continue
        }

        var fsTypeBuf [64]uint16
        err := windows.GetVolumeInformation(
            rootPtr,
            nil, 0,  // volume name (not needed)
            nil,     // serial number (not needed)
            nil,     // max component length (not needed)
            nil,     // filesystem flags (not needed)
            &fsTypeBuf[0],
            uint32(len(fsTypeBuf)),
        )
        if err != nil {
            continue // no media / not ready
        }

        drives = append(drives, drive.Drive{
            Device:     `\\.\` + letter + `:`,
            MountPoint: rootPath,
            FSType:     windows.UTF16ToString(fsTypeBuf[:]),
        })
    }

    return drives, nil
}


