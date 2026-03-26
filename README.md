# Concat Video Tool

A command-line tool for concatenating MP4 video files from external drives (USB drives, external HDDs) on Linux systems. This tool automatically detects mounted external drives, organizes video files by date, and creates concatenated video files using FFmpeg.
## Todo
- [ ] Test on windows machine
- [ ] Test on mac
## Features

- **Automatic Drive Detection**: Scans for mounted external drives on Linux systems
- **Date-based Organization**: Groups video files by year, month, day, and hour
- **Parallel Processing**: Uses goroutines for efficient concatenation of multiple video sets
- **FFmpeg Integration**: Leverages FFmpeg for high-quality video concatenation


## Prerequisites

- FFmpeg installed on your system
- Linux operating system (for drive detection)

## Installation

### Option 1: Download Release (Recommended)

1. Go to the [Releases](https://github.com/mosesmuiru/concat/releases) page
2. Download the latest release for your platform (Linux)
3. Extract the `cli` binary to a directory in your PATH, or run it directly

### Option 2: Build from Source

For developers or users who prefer to build from source:

1. Ensure you have Go 1.26.1 or later installed
2. Clone the repository:
```bash
git clone https://github.com/mosesmuiru/concat.git
cd concat
```

3. Build the executable:
```bash
go build -o cli
```

## Usage

### Basic Usage

```bash
./cli -l <device_id>
```

Where:
- `-l, --linux`: Enable Linux mode for drive detection
- `<device_id>`: The device UUID or identifier for the video files

### Example

```bash
./cli -l abc123-def456-ghi789
```

## How It Works

1. **Drive Detection**: The tool scans mounted external drives looking for devices mounted under `/media`, `/mnt`, or `/run/media` directories.

2. **File Discovery**: Searches for video files in the path structure:
   ```
   /mount/point/ex_nvr/<device_id>/hi_quality/
   ```

3. **File Organization**: Groups MP4 files by date extracted from directory paths matching the pattern:
   ```
   .../ex_nvr/[uuid]/[category]/[year]/[month]/[day]/[hour]
   ```

4. **Concatenation**: Creates text files listing video files for each hour, then uses FFmpeg to concatenate them into single video files.

5. **Output**: Saves concatenated videos in date-organized directories within your current working directory.

## Directory Structure

The tool expects source files organized as:
```
/media/external-drive/ex_nvr/
├── <device_id>/
│   └── hi_quality/
│       ├── category1/
│       │   ├── 2024/
│       │   │   ├── 01/
│       │   │   │   ├── 15/
│       │   │   │   │   ├── 10/
|       │   |   │   │   │   │   ├── 107958935646/ # Contains MP4 files in 1 mins format
|       │       │   │   │   │   └── 1054358960798/  # Contains MP4 files in 1 mins format
│       │   │   │   └── 16/
│       │   │   │   │   ├── 11/
|       │   |   │   │   │   │   ├── 117958935646/ # Contains MP4 files in 1 mins format
|       │       │   │   │   │   └── 1154358960798/  # Contains MP4 files in 1 mins format
│       │   │   └── 02/
│       │   └── ...
│       └── category2/
│           └── ...
```

Output files are saved as:
```
./
├── 2024/
│   ├── 01/
│   │   ├── 15/
│   │   │   ├── 10.mp4  # Concatenated video for 10 AM
│   │   │   └── 11.mp4  # Concatenated video for 11 AM
│   │   └── 16/
│   └── 02/
└── ...
```

## Dependencies

- **github.com/spf13/cobra**: For CLI argument parsing
- **FFmpeg**: For video concatenation (must be installed separately)

## Installation of FFmpeg

### Ubuntu/Debian:
```bash
sudo apt update
sudo apt install ffmpeg
```

### CentOS/RHEL:
```bash
sudo yum install ffmpeg
```

### macOS (with Homebrew):
```bash
brew install ffmpeg
```

## Troubleshooting

- **No external drives found**: Ensure your external drives are properly mounted
- **Permission errors**: Run with appropriate permissions or mount drives with correct access rights
- **FFmpeg not found**: Install FFmpeg as shown above
- **No files to concatenate**: Check that your drive contains files in the expected directory structure

