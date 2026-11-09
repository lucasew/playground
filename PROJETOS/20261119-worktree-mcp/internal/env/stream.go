package env

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Stream reads the output of one exec as whole lines at a line offset.
// The file is append-only: writers only append, and readers never rewind it.
type Stream struct {
	mu   sync.Mutex
	path string
}

// OpenStream reads the stream stored at path.
func OpenStream(path string) *Stream {
	return &Stream{path: path}
}

// Page is one read of a stream.
type Page struct {
	Lines         []string `json:"lines"`
	NextOffset    int      `json:"next_offset"`
	State         State    `json:"state"`
	ExitCode      int      `json:"exit_code"`
	TruncatedFrom int      `json:"truncated_from"`
}

// State is the lifecycle of one exec.
type State string

const (
	Running State = "running"
	Exited  State = "exited"
)

// Read returns whole lines from offset, up to limit lines. An empty pattern
// keeps every line. The returned NextOffset counts every scanned line,
// whether it matched. The state is reported as running; the caller supplies
// the exit status from its own record.
func (s *Stream) Read(offset, limit int, pattern string) (Page, error) {
	if offset < 0 {
		return Page{}, fmt.Errorf("%w: %d", ErrOffset, offset)
	}
	if limit < 0 {
		limit = 0
	}
	var re *regexp.Regexp
	if pattern != "" {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return Page{}, fmt.Errorf("%w: %w", ErrPattern, err)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		return Page{}, err
	}
	defer f.Close()
	page := Page{Lines: []string{}, NextOffset: offset}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for n := 0; scanner.Scan(); n++ {
		if n < offset {
			continue
		}
		line := scanner.Text()
		if re == nil || re.MatchString(line) {
			page.Lines = append(page.Lines, line)
		}
		page.NextOffset = n + 1
		if limit > 0 && len(page.Lines) >= limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return Page{}, err
	}
	return page, nil
}

// Offset is the number of lines currently in the stream.
func (s *Stream) Offset() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	n := 0
	for scanner.Scan() {
		n++
	}
	return n, scanner.Err()
}

// StripANSI removes terminal escape sequences from one line.
func StripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) {
			j := i + 1
			if s[j] == '[' {
				j++
				for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
					j++
				}
				i = j
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
