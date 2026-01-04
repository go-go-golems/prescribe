package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-go-golems/prescribe/internal/tokens"
	"github.com/pkg/errors"
)

const (
	gitContextDefaultMaxBytes  = 120_000
	gitContextDefaultMaxTokens = 2_000
)

func truncateWithCaps(s string, maxBytes, maxTokens int) (string, bool) {
	original := s

	if maxTokens > 0 {
		// Best-effort: binary search the largest prefix under maxTokens.
		if tokens.Count(s) > maxTokens {
			runes := []rune(s)
			lo, hi := 0, len(runes)
			best := 0
			for lo <= hi {
				mid := (lo + hi) / 2
				if mid < 0 {
					break
				}
				cand := string(runes[:mid])
				if tokens.Count(cand) <= maxTokens {
					best = mid
					lo = mid + 1
				} else {
					hi = mid - 1
				}
			}
			s = string(runes[:best])
		}
	}

	if maxBytes > 0 && len(s) > maxBytes {
		s = s[:maxBytes]
	}

	if s == original {
		return s, false
	}

	marker := fmt.Sprintf("\n... [TRUNCATED: max_bytes=%d max_tokens=%d]\n", maxBytes, maxTokens)
	return strings.TrimRight(s, "\n") + marker, true
}

type commitHeader struct {
	SHA     string
	Author  string
	Date    string
	Subject string
}

func (s *Service) getCommitHeader(ref string) (commitHeader, error) {
	cmd := exec.Command("git", "show", "-s", "--date=iso-strict", "--format=%H%x1f%an%x1f%ad%x1f%s", ref)
	cmd.Dir = s.repoPath
	out, err := cmd.Output()
	if err != nil {
		return commitHeader{}, errors.Wrap(err, "failed to read commit header")
	}
	fields := strings.Split(strings.TrimSpace(string(out)), string([]byte{0x1f}))
	if len(fields) < 4 {
		return commitHeader{}, fmt.Errorf("unexpected commit header format for %q", ref)
	}
	return commitHeader{
		SHA:     strings.TrimSpace(fields[0]),
		Author:  strings.TrimSpace(fields[1]),
		Date:    strings.TrimSpace(fields[2]),
		Subject: strings.TrimSpace(fields[3]),
	}, nil
}

type numstatLine struct {
	Path      string
	Additions int
	Deletions int
}

func (s *Service) getCommitNumstat(ref string) ([]numstatLine, error) {
	// diff-tree yields machine-friendly numstat without patch output.
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--numstat", "-r", ref)
	cmd.Dir = s.repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read commit numstat")
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	stats := make([]numstatLine, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		additions := 0
		deletions := 0
		if parts[0] != "-" {
			if v, err := strconv.Atoi(parts[0]); err == nil {
				additions = v
			}
		}
		if parts[1] != "-" {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				deletions = v
			}
		}
		path := strings.Join(parts[2:], "\t")
		stats = append(stats, numstatLine{Path: path, Additions: additions, Deletions: deletions})
	}
	return stats, nil
}

func (s *Service) BuildCommitMetadataContext(ref string, includeNumstat bool) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", nil
	}

	hdr, err := s.getCommitHeader(ref)
	if err != nil {
		return "", err
	}

	var (
		filesChanged int
		additions    int
		deletions    int
		perFile      []numstatLine
	)
	numstat, err := s.getCommitNumstat(hdr.SHA)
	if err == nil {
		seen := map[string]bool{}
		for _, ns := range numstat {
			additions += ns.Additions
			deletions += ns.Deletions
			if !seen[ns.Path] {
				seen[ns.Path] = true
				filesChanged++
			}
		}
		if includeNumstat {
			perFile = numstat
		}
	}

	shortSHA := hdr.SHA
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<git_commit ref=\"%s\" sha=\"%s\">\n", xmlEscapeAttr(ref), xmlEscapeAttr(shortSHA)))
	b.WriteString(fmt.Sprintf("<subject>%s</subject>\n", xmlEscapeText(hdr.Subject)))
	b.WriteString(fmt.Sprintf("<author>%s</author>\n", xmlEscapeText(hdr.Author)))
	b.WriteString(fmt.Sprintf("<date>%s</date>\n", xmlEscapeText(hdr.Date)))
	b.WriteString(fmt.Sprintf("<summary files=\"%d\" additions=\"%d\" deletions=\"%d\"/>\n", filesChanged, additions, deletions))
	if includeNumstat && len(perFile) > 0 {
		var nsb strings.Builder
		for _, ns := range perFile {
			nsb.WriteString(fmt.Sprintf(
				"<file path=\"%s\" additions=\"%d\" deletions=\"%d\"/>\n",
				xmlEscapeAttr(ns.Path),
				ns.Additions,
				ns.Deletions,
			))
		}
		numstatBody, _ := truncateWithCaps(nsb.String(), gitContextDefaultMaxBytes, gitContextDefaultMaxTokens)
		b.WriteString("<numstat>\n")
		b.WriteString(strings.TrimRight(numstatBody, "\n"))
		b.WriteString("\n")
		b.WriteString("</numstat>\n")
	}
	b.WriteString("</git_commit>\n")

	return b.String(), nil
}

func (s *Service) BuildCommitPatchContext(ref string, paths []string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", nil
	}

	hdr, err := s.getCommitHeader(ref)
	if err != nil {
		return "", err
	}

	args := []string{"show", "--format=", "--patch", hdr.SHA}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = s.repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to read commit patch")
	}
	patch, _ := truncateWithCaps(string(out), gitContextDefaultMaxBytes, gitContextDefaultMaxTokens)

	shortSHA := hdr.SHA
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<git_commit_patch ref=\"%s\" sha=\"%s\"", xmlEscapeAttr(ref), xmlEscapeAttr(shortSHA)))
	if len(paths) > 0 {
		b.WriteString(fmt.Sprintf(" paths=\"%s\"", xmlEscapeAttr(strings.Join(paths, ","))))
	}
	b.WriteString(">\n")
	b.WriteString("<patch>\n")
	b.WriteString(strings.TrimRight(patch, "\n"))
	b.WriteString("\n</patch>\n</git_commit_patch>\n")
	return b.String(), nil
}

func (s *Service) BuildFileAtRefContext(ref, filePath string) (string, error) {
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(filePath) == "" {
		return "", nil
	}

	content, err := s.GetFileContent(ref, filePath)
	if err != nil {
		return "", err
	}
	content, _ = truncateWithCaps(content, gitContextDefaultMaxBytes, gitContextDefaultMaxTokens)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<git_file_at_ref ref=\"%s\" path=\"%s\">\n", xmlEscapeAttr(ref), xmlEscapeAttr(filePath)))
	b.WriteString("<content>\n")
	b.WriteString(strings.TrimRight(content, "\n"))
	b.WriteString("\n</content>\n</git_file_at_ref>\n")
	return b.String(), nil
}

func (s *Service) BuildFileDiffContext(fromRef, toRef, filePath string) (string, error) {
	if strings.TrimSpace(fromRef) == "" || strings.TrimSpace(toRef) == "" || strings.TrimSpace(filePath) == "" {
		return "", nil
	}

	cmd := exec.Command("git", "diff", fromRef, toRef, "--", filePath)
	cmd.Dir = s.repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to read file diff")
	}
	diff, _ := truncateWithCaps(string(out), gitContextDefaultMaxBytes, gitContextDefaultMaxTokens)

	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"<git_file_diff from=\"%s\" to=\"%s\" path=\"%s\">\n",
		xmlEscapeAttr(fromRef),
		xmlEscapeAttr(toRef),
		xmlEscapeAttr(filePath),
	))
	b.WriteString("<diff>\n")
	b.WriteString(strings.TrimRight(diff, "\n"))
	b.WriteString("\n</diff>\n</git_file_diff>\n")
	return b.String(), nil
}
