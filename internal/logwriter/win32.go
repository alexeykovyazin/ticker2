//go:build windows

package logwriter

import (
	"bytes"
	"fmt"

	"golang.org/x/sys/windows"
)

type Win32Log struct {
	handle windows.Handle
}

func OpenWin32(path string) (*Win32Log, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("convert path: %w", err)
	}

	flags := uint32(windows.FILE_FLAG_OVERLAPPED | windows.FILE_FLAG_WRITE_THROUGH)
	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_ALWAYS,
		flags,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateFile %q: %w", path, err)
	}

	return &Win32Log{handle: handle}, nil
}

func (w *Win32Log) Close() error {
	if w.handle == windows.InvalidHandle {
		return nil
	}
	err := windows.CloseHandle(w.handle)
	w.handle = windows.InvalidHandle
	return err
}

func (w *Win32Log) fileSize() (int64, error) {
	var high int32
	low, err := windows.SetFilePointer(w.handle, 0, &high, windows.FILE_END)
	if err != nil {
		return 0, fmt.Errorf("SetFilePointer FILE_END: %w", err)
	}
	return int64(high)<<32 | int64(uint32(low)), nil
}

func (w *Win32Log) setFilePointer(offset int64) error {
	var high int32 = int32(offset >> 32)
	low := int32(offset & 0xffffffff)
	_, err := windows.SetFilePointer(w.handle, low, &high, windows.FILE_BEGIN)
	if err != nil {
		return fmt.Errorf("SetFilePointer at %d: %w", offset, err)
	}
	return nil
}

func (w *Win32Log) createOverlapped(offset int64) (*windows.Overlapped, error) {
	ev, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateEvent: %w", err)
	}

	ov := &windows.Overlapped{
		HEvent: ev,
	}
	ov.Offset = uint32(offset & 0xffffffff)
	ov.OffsetHigh = uint32(offset >> 32)
	return ov, nil
}

func (w *Win32Log) waitOverlapped(ov *windows.Overlapped, done *uint32) error {
	defer windows.CloseHandle(ov.HEvent)

	err := windows.GetOverlappedResult(w.handle, ov, done, true)
	if err != nil {
		return fmt.Errorf("GetOverlappedResult: %w", err)
	}
	return nil
}

func (w *Win32Log) writeAt(offset int64, data []byte) (uint32, error) {
	ov, err := w.createOverlapped(offset)
	if err != nil {
		return 0, err
	}

	var written uint32
	err = windows.WriteFile(w.handle, data, &written, ov)
	if err == windows.ERROR_IO_PENDING {
		err = w.waitOverlapped(ov, &written)
	} else {
		windows.CloseHandle(ov.HEvent)
	}
	if err != nil {
		return 0, fmt.Errorf("WriteFile at %d: %w", offset, err)
	}
	return written, nil
}

func (w *Win32Log) readAt(offset int64, buf []byte) (uint32, error) {
	ov, err := w.createOverlapped(offset)
	if err != nil {
		return 0, err
	}

	var read uint32
	err = windows.ReadFile(w.handle, buf, &read, ov)
	if err == windows.ERROR_IO_PENDING {
		err = w.waitOverlapped(ov, &read)
	} else {
		windows.CloseHandle(ov.HEvent)
	}
	if err != nil {
		return 0, fmt.Errorf("ReadFile at %d: %w", offset, err)
	}
	return read, nil
}

func (w *Win32Log) AppendLine(line string) error {
	data := []byte(line)

	offset, err := w.fileSize()
	if err != nil {
		return err
	}

	written, err := w.writeAt(offset, data)
	if err != nil {
		return err
	}

	newEOF := offset + int64(written)
	if err := w.setFilePointer(newEOF); err != nil {
		return err
	}
	if err := windows.SetEndOfFile(w.handle); err != nil {
		return fmt.Errorf("SetEndOfFile: %w", err)
	}
	if err := windows.FlushFileBuffers(w.handle); err != nil {
		return fmt.Errorf("FlushFileBuffers: %w", err)
	}

	readBuf := make([]byte, len(data))
	read, err := w.readAt(offset, readBuf)
	if err != nil {
		return fmt.Errorf("verify read: %w", err)
	}
	if read != uint32(len(data)) || !bytes.Equal(readBuf[:read], data) {
		return fmt.Errorf("verify read: got %q, want %q", readBuf[:read], data)
	}

	return nil
}
