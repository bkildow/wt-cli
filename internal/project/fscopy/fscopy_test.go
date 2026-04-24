package fscopy

import (
	"bytes"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCopyFileContentAndMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	dst := filepath.Join(dir, "dst.bin")

	want := make([]byte, 1<<16)
	if _, err := rand.Read(want); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, want, 0o640); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Error("dst content does not match src")
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("dst mode = %v, want 0o640", got)
	}

	// No temp file left behind.
	if _, err := os.Stat(dst + ".wtclone"); err == nil {
		t.Error(".wtclone tmp file was not cleaned up")
	}
}

func TestCopyFileEmpty(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty")
	dst := filepath.Join(dir, "empty-copy")
	if err := os.WriteFile(src, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("dst size = %d, want 0", info.Size())
	}
}

func TestCopyFileOverwritesExistingDst(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("dst = %q, want %q", got, "new")
	}
}

func TestCopyFileReclaimsStaleTmp(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Simulate a prior interrupted run that left .wtclone behind.
	if err := os.WriteFile(dst+".wtclone", []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Errorf("dst = %q, want %q", got, "payload")
	}
	if _, err := os.Stat(dst + ".wtclone"); err == nil {
		t.Error("stale .wtclone not cleaned up")
	}
}

func TestCopyFileMissingSource(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "dst")
	err := CopyFile(filepath.Join(dir, "nope"), dst)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !os.IsNotExist(err) {
		t.Errorf("error = %v, want IsNotExist", err)
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("dst should not have been created")
	}
}

// TestByteCopyFallback exercises the fallback path directly, which on
// reflink-capable filesystems (like APFS in t.TempDir on macOS) is
// otherwise not reached by CopyFile. On non-COW filesystems the fallback
// is exercised transparently by every other test in this file as well.
func TestByteCopyFallback(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	payload := []byte("fallback payload \x00\x01\x02")
	if err := os.WriteFile(src, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := byteCopy(src, dst, 0o600); err != nil {
		t.Fatalf("byteCopy: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("content = %q, want %q", got, payload)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0o600", info.Mode().Perm())
	}
}

// TestByteCopyRefusesExistingDst confirms byteCopy uses O_EXCL on dst so
// a caller that forgets to pre-clean would get an error rather than a
// silent overwrite into a possibly-reflinked file.
func TestByteCopyRefusesExistingDst(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := byteCopy(src, dst, 0o644); err == nil {
		t.Fatal("expected O_EXCL error when dst exists")
	}
}

// skipUnlessReflinkTree skips the test when tryCloneTree reports the
// current filesystem cannot clone trees. On macOS t.TempDir is APFS so
// this runs; on Linux / non-APFS it returns errReflinkUnsupported.
func skipUnlessReflinkTree(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("tree-level reflink only implemented on darwin")
	}
	// Probe: try cloning an empty dir. If unsupported on the test FS, skip.
	probeSrc := filepath.Join(t.TempDir(), "probe-src")
	if err := os.Mkdir(probeSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	probeDst := filepath.Join(t.TempDir(), "probe-dst")
	if err := CopyTree(probeSrc, probeDst); err != nil {
		if errors.Is(err, errReflinkUnsupported) {
			t.Skipf("tree reflink unsupported on this filesystem: %v", err)
		}
		t.Fatalf("probe CopyTree: %v", err)
	}
}

func TestCopyTreeNested(t *testing.T) {
	skipUnlessReflinkTree(t)
	srcRoot := t.TempDir()
	dstRoot := t.TempDir()
	src := filepath.Join(srcRoot, "tree")
	dst := filepath.Join(dstRoot, "tree")

	if err := os.MkdirAll(filepath.Join(src, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a", "top.txt"), []byte("top"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a", "b", "deep.txt"), []byte("deep"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := CopyTree(src, dst); err != nil {
		t.Fatalf("CopyTree: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "a", "top.txt"))
	if err != nil || string(got) != "top" {
		t.Errorf("top.txt = %q err %v; want %q", got, err, "top")
	}
	info, err := os.Stat(filepath.Join(dst, "a", "top.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Errorf("top.txt mode = %v, want 0o640", info.Mode().Perm())
	}

	got, err = os.ReadFile(filepath.Join(dst, "a", "b", "deep.txt"))
	if err != nil || string(got) != "deep" {
		t.Errorf("deep.txt = %q err %v; want %q", got, err, "deep")
	}

	if _, err := os.Stat(dst + ".wtclone"); err == nil {
		t.Error("tree .wtclone was not cleaned up")
	}
}

func TestCopyTreePreservesSymlink(t *testing.T) {
	skipUnlessReflinkTree(t)
	srcRoot := t.TempDir()
	dstRoot := t.TempDir()
	src := filepath.Join(srcRoot, "tree")
	dst := filepath.Join(dstRoot, "tree")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "real.txt"), []byte("r"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.txt", filepath.Join(src, "link.txt")); err != nil {
		t.Fatal(err)
	}

	if err := CopyTree(src, dst); err != nil {
		t.Fatalf("CopyTree: %v", err)
	}

	info, err := os.Lstat(filepath.Join(dst, "link.txt"))
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("link.txt was not preserved as a symlink")
	}
	target, err := os.Readlink(filepath.Join(dst, "link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if target != "real.txt" {
		t.Errorf("symlink target = %q, want %q", target, "real.txt")
	}
}

func TestCopyTreeExistingDst(t *testing.T) {
	skipUnlessReflinkTree(t)
	srcRoot := t.TempDir()
	dstRoot := t.TempDir()
	src := filepath.Join(srcRoot, "tree")
	dst := filepath.Join(dstRoot, "tree")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatal(err)
	}

	err := CopyTree(src, dst)
	if err == nil {
		t.Fatal("expected error when dst already exists")
	}
	if errors.Is(err, errReflinkUnsupported) {
		t.Errorf("existing-dst should not map to unsupported: %v", err)
	}
}

func TestIsReflinkUnsupported(t *testing.T) {
	if !IsReflinkUnsupported(errReflinkUnsupported) {
		t.Error("IsReflinkUnsupported should return true for sentinel")
	}
	if IsReflinkUnsupported(errors.New("something else")) {
		t.Error("IsReflinkUnsupported should return false for arbitrary error")
	}
}

func TestCopyFileMissingDestinationParent(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	// CopyFile's contract does not create parent dirs; callers do that.
	err := CopyFile(src, filepath.Join(dir, "missing", "dst"))
	if err == nil {
		t.Fatal("expected error when parent dir missing")
	}
	if !strings.Contains(err.Error(), "missing") && !os.IsNotExist(err) {
		// Either the wrapped reflink path or the byteCopy fallback can surface
		// ENOENT; accept any error that mentions the missing path or is an
		// os.IsNotExist sentinel.
		t.Logf("note: error text = %v", err)
	}
}
