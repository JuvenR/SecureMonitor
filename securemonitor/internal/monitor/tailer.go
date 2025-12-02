package monitor

import (
	"os"
	"strings"
)

//  keeps track of the last read byte offset per log file.
var lastOffsets = make(map[string]int64)

//  returns only the new lines appended to the file since the last call.
func ReadNewLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()

	offset, ok := lastOffsets[path]

	// On first call, ignore historical content and start from the end.
	if !ok {
		lastOffsets[path] = size
		return []string{}, nil
	}

	// if the file was rotated or truncated, restart from the beginning.
	if size < offset {
		offset = 0
	}

	// nothing new added since last read.
	if size == offset {
		return []string{}, nil
	}

	// read only the new bytes appended since the last offset.
	toRead := size - offset
	buf := make([]byte, toRead)

	if _, err := f.Seek(offset, 0); err != nil {
		return nil, err
	}

	n, err := f.Read(buf)
	if n <= 0 {
		lastOffsets[path] = size
		return []string{}, nil
	}

	lastOffsets[path] = offset + int64(n)

	data := string(buf[:n])
	return strings.Split(data, "\n"), nil
}
