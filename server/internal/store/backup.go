package store

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const keepBackups = 30

// BackupInfo is one backup file on disk.
type BackupInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Time int64  `json:"time"`
}

func (s *Store) backupDir() string { return filepath.Join(s.dir, "backups") }

// Backup writes a consistent copy of the database, then prunes old copies.
// extraDir is an optional second location such as a USB stick; a missing stick
// must not stop the main backup, so its error is returned after the local copy exists.
func (s *Store) Backup(ctx context.Context, extraDir string) error {
	if err := os.MkdirAll(s.backupDir(), 0o755); err != nil {
		return err
	}
	name := "amk-bar-" + s.now().Local().Format("2006-01-02-150405") + ".db"
	target := filepath.Join(s.backupDir(), name)
	escaped := strings.ReplaceAll(filepath.ToSlash(target), "'", "''")
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO '"+escaped+"'"); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	if err := prune(s.backupDir()); err != nil {
		return err
	}
	if extraDir = strings.TrimSpace(extraDir); extraDir != "" {
		if err := copyFile(target, filepath.Join(extraDir, name)); err != nil {
			return fmt.Errorf("copy to second folder: %w", err)
		}
		return prune(extraDir)
	}
	return nil
}

// BackedUpToday reports whether a backup was made on the current calendar day.
func (s *Store) BackedUpToday() bool {
	stamp := "amk-bar-" + s.now().Local().Format("2006-01-02")
	entries, err := os.ReadDir(s.backupDir())
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), stamp) {
			return true
		}
	}
	return false
}

func (s *Store) Backups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.backupDir())
	if os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []BackupInfo{}
	for _, e := range entries {
		if !isBackup(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, BackupInfo{Name: e.Name(), Size: info.Size(), Time: info.ModTime().UnixMilli()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

func isBackup(name string) bool {
	return strings.HasPrefix(name, "amk-bar-") && strings.HasSuffix(name, ".db")
}

func prune(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		if isBackup(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for len(names) > keepBackups {
		if err := os.Remove(filepath.Join(dir, names[0])); err != nil {
			return err
		}
		names = names[1:]
	}
	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
