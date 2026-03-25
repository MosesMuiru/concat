package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mosesmuiru/concat/devices"
	"github.com/mosesmuiru/concat/devices/drive"
	"github.com/spf13/cobra"
)



var (
	win bool
	linux   bool
)

func init() {
	rootCmd.Flags().BoolVarP(&win, "windows", "w", false, "windows")
	rootCmd.Flags().BoolVarP(&linux, "linux", "l", false, "linux")
}
//

// // --- Linux: parse /proc/mounts ---
// func getLinuxDrives() ([]drive.Drive, error) {
// 	f, err := os.Open("/proc/mounts")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()

// 	var drives []drive.Drive

// 	scanner := bufio.NewScanner(f)
// 	for scanner.Scan() {
// 		fields := strings.Fields(scanner.Text())
// 		if len(fields) < 3 {
// 			continue
// 		}
// 		device, mountPoint, fsType := fields[0], fields[1], fields[2]

// 		// USB drives typically mount under /media or /mnt, on /dev/sd* or /dev/nvme*
// 		isExternal := (strings.HasPrefix(device, "/dev/sd") ||
// 			strings.HasPrefix(device, "/dev/nvme")) &&
// 			(strings.HasPrefix(mountPoint, "/media") ||
// 				strings.HasPrefix(mountPoint, "/mnt") ||
// 				strings.HasPrefix(mountPoint, "/run/media"))

// 		if isExternal {
// 			drives = append(drives, drive.Drive{device, mountPoint, fsType})
// 		}
// 	}
// 	return drives, scanner.Err()
// }


// can now read the files in that directory

// getAllFiles walks a mount point and returns all file paths
func getAllFiles(mountPoint string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(mountPoint, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Permission errors etc. — skip but don't stop
			fmt.Fprintf(os.Stderr, "  skipping %s: %v\n", path, err)
			return nil
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}


func nameFile(dir string) (*string, error) {
	re := regexp.MustCompile(`.*/ex_nvr/([a-f0-9\-]+)/([a-z_]+)/(\d{4})/(\d{2})/(\d{2})/(\d{2})$`)

	match := re.FindStringSubmatch(dir)

	if match != nil {

		//  Correct format: year_month_day_hour.txt
		filename := fmt.Sprintf("%s_%s_%s_%s.txt",
			match[3], // year
			match[4], // month
			match[5], // day
			match[6], // hour
		)

		filename = filepath.Join("mp4TextFiles", filename)
		err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
		if err != nil {
			panic(err)
		}

		return &filename, nil
	}

	return nil, errors.New("Invalid")
}

func createTextForMP4File(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func writeDir(dirPath string) (string, error) {
	var lastVistedDir string
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}


		filename, _ := nameFile(path)

		if info.IsDir() && filename != nil {

			lastVistedDir = path
		}

		if !info.IsDir() {

			dir, _ := nameFile(lastVistedDir)
			file, err := os.OpenFile(*dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				println("failed to create a file")
				panic(err)
			}
			defer file.Close()

			// this writes the mp4s to the file
			if filepath.Ext(path) == ".mp4" {
				_, err = file.WriteString("file '" + path + "'\n")
				if err != nil {

			println("failed a file")
					panic(err)
				}

			}

		}

		return nil
	})

	return lastVistedDir, err

}

func concatWithFFMpeg(input string, output string) error {

	cmd := exec.Command(
		"ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", input,
		"-c", "copy",
		output, "-n",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func concatenate(drives []drive.Drive, device_uuid string) {

	for _, drive := range drives {

		
		usbPath := drive.MountPoint + "/ex_nvr/" + device_uuid + "/hi_quality/"
		dirs, _ := os.ReadDir(usbPath)

		for _, dir := range dirs {

			writeDir(usbPath + dir.Name())
		}

		maxWorkers := 4

		txtFiles, err := os.ReadDir("mp4TextFiles")
	
		if err != nil {
			println("Sorry no files to concatinate found")
		}

		jobs := make(chan string, len(txtFiles))

		var wg sync.WaitGroup

		for i := 0; i < maxWorkers; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				for txtFile := range jobs {

					dir := filepath.Dir(txtFile)
					day := strings.TrimSuffix(filepath.Base(txtFile), ".txt") // "2020_12_01"
					parts := strings.Split(day, "_")
					dirPath := filepath.Join(append([]string{dir}, parts[:3]...)...)

					err := os.MkdirAll(dirPath, os.ModePerm)
					if err != nil {
						panic(err)
					}
					outputFile := filepath.Join(dirPath, parts[len(parts)-1]+".mp4")
			
			println("txxxxtt files", dir)

					if err := concatWithFFMpeg("mp4TextFiles/" + txtFile, outputFile); err != nil {
						fmt.Printf("Error processing %s: %v\n", txtFile)
					} else {
						fmt.Printf("Done: %s\n", txtFile)
					}
				}
			}()
		}
		for _, f := range txtFiles {
			jobs <- f.Name()
		}
		close(jobs)

		wg.Wait()
		fmt.Println("All done.")
		err = os.RemoveAll("mp4TextFiles")
		if err != nil {
			fmt.Println("Error:", err)
		}

	}

}

var rootCmd = &cobra.Command{
	Use:   "cli <device_id>",
	Short: "Returns the given device ID",
	Args:  cobra.ExactArgs(1),
	Run:   runConcatinator,
}

func runConcatinator(cmd *cobra.Command, args []string) {
	var text string

	if len(args) > 0 {
		text = strings.Join(args, " ")
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		text = strings.Join(lines, "\n")
	}

	if text == "" {
		fmt.Println("No input text provided")
		return
	}

	result := text
	if linux {
		// get mounted devices
		drives, err := devices.GetLinuxDrives()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting drives: %v\n", err)
			os.Exit(1)
		}

		if len(drives) == 0 {
			fmt.Println("No external drives found.")
			return
		}
		// concatinate
		concatenate(drives, result)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// notes - use go routines
func main() {
	Execute()

}
