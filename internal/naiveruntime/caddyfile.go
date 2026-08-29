package naiveruntime

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

type Inspection struct {
	Usernames       []string `json:"usernames"`
	CredentialCount int      `json:"credential_count"`

	credentials     []parsedCredential
	credentialStart int
	credentialEnd   int
	indent          string
	newline         string
}

type parsedCredential struct {
	username string
	password string
}

type tokenKind uint8

const (
	tokenWord tokenKind = iota
	tokenQuoted
	tokenLBrace
	tokenRBrace
)

type token struct {
	kind        tokenKind
	value       string
	start       int
	end         int
	line        int
	depth       int
	firstOnLine bool
}

type credentialLine struct {
	username string
	password string
	start    int
	end      int
	indent   string
	newline  string
}

func InspectCaddyfile(input []byte) (Inspection, error) {
	tokens, err := lexCaddyfile(input)
	if err != nil {
		return Inspection{}, err
	}

	openIndex, closeIndex, err := findForwardProxyBlock(tokens)
	if err != nil {
		return Inspection{}, err
	}

	lines, err := parseCredentialLines(input, tokens, openIndex, closeIndex)
	if err != nil {
		return Inspection{}, err
	}
	if len(lines) == 0 {
		return Inspection{}, errors.New("naiveruntime: forward_proxy contains no basic_auth credentials")
	}

	for i := 1; i < len(lines); i++ {
		if lines[i-1].end != lines[i].start {
			return Inspection{}, errors.New("naiveruntime: basic_auth directives are not contiguous")
		}
		if lines[i].newline != lines[0].newline {
			return Inspection{}, errors.New("naiveruntime: mixed credential line endings are ambiguous")
		}
	}

	inspection := Inspection{
		Usernames:       make([]string, 0, len(lines)),
		CredentialCount: len(lines),
		credentials:     make([]parsedCredential, 0, len(lines)),
		credentialStart: lines[0].start,
		credentialEnd:   lines[len(lines)-1].end,
		indent:          lines[0].indent,
		newline:         lines[0].newline,
	}
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		if _, ok := seen[line.username]; ok {
			return Inspection{}, fmt.Errorf("naiveruntime: duplicate basic_auth username %q", line.username)
		}
		seen[line.username] = struct{}{}
		inspection.Usernames = append(inspection.Usernames, line.username)
		inspection.credentials = append(inspection.credentials, parsedCredential{username: line.username, password: line.password})
	}
	return inspection, nil
}

func RenderCredentials(input []byte, credentials []runtimecred.DesiredCredential) ([]byte, error) {
	inspection, err := InspectCaddyfile(input)
	if err != nil {
		return nil, err
	}
	if len(credentials) == 0 {
		return nil, errors.New("naiveruntime: credential set must not be empty")
	}

	seen := make(map[string]struct{}, len(credentials))
	active := make([]runtimecred.DesiredCredential, 0, len(credentials))
	for _, credential := range credentials {
		if err := runtimecred.ValidateUsername(credential.Username); err != nil {
			return nil, fmt.Errorf("naiveruntime: invalid username: %w", err)
		}
		if _, ok := seen[credential.Username]; ok {
			return nil, fmt.Errorf("naiveruntime: duplicate username %q", credential.Username)
		}
		seen[credential.Username] = struct{}{}

		switch credential.Status {
		case runtimecred.CredentialActive, runtimecred.CredentialDisabled, runtimecred.CredentialRevoked:
		default:
			return nil, fmt.Errorf("naiveruntime: invalid credential status %q", credential.Status)
		}
		if err := runtimecred.ValidatePassword(credential.Password(), true); err != nil {
			return nil, fmt.Errorf("naiveruntime: invalid password for %q: %w", credential.Username, err)
		}
		if credential.Status == runtimecred.CredentialActive {
			active = append(active, credential)
		}
	}
	if len(active) == 0 {
		return nil, errors.New("naiveruntime: at least one active credential is required")
	}
	if equivalentActiveCredentials(active, inspection.credentials) {
		return append([]byte(nil), input...), nil
	}

	var replacement strings.Builder
	for _, credential := range active {
		replacement.WriteString(inspection.indent)
		replacement.WriteString("basic_auth ")
		replacement.WriteString(credential.Username)
		replacement.WriteByte(' ')
		encodedPassword, err := quoteCaddyToken(credential.Password())
		if err != nil {
			return nil, fmt.Errorf("naiveruntime: render password for %q: %w", credential.Username, err)
		}
		replacement.WriteString(encodedPassword)
		replacement.WriteString(inspection.newline)
	}

	output := make([]byte, 0, len(input)-inspection.credentialEnd+inspection.credentialStart+replacement.Len())
	output = append(output, input[:inspection.credentialStart]...)
	output = append(output, replacement.String()...)
	output = append(output, input[inspection.credentialEnd:]...)
	return output, nil
}

func equivalentActiveCredentials(active []runtimecred.DesiredCredential, live []parsedCredential) bool {
	if len(active) != len(live) {
		return false
	}
	passwords := make(map[string]string, len(active))
	for _, credential := range active {
		passwords[credential.Username] = credential.Password()
	}
	for _, credential := range live {
		password, ok := passwords[credential.username]
		if !ok || password != credential.password {
			return false
		}
	}
	return true
}

func findForwardProxyBlock(tokens []token) (int, int, error) {
	var openings []int
	for i, tok := range tokens {
		if tok.kind != tokenWord || tok.value != "forward_proxy" || !tok.firstOnLine {
			continue
		}
		if i+1 >= len(tokens) || tokens[i+1].kind != tokenLBrace || tokens[i+1].line != tok.line {
			return 0, 0, errors.New("naiveruntime: malformed forward_proxy directive")
		}
		openings = append(openings, i+1)
	}
	if len(openings) != 1 {
		return 0, 0, fmt.Errorf("naiveruntime: expected exactly one forward_proxy block, found %d", len(openings))
	}

	open := openings[0]
	openDepth := tokens[open].depth
	for i := open + 1; i < len(tokens); i++ {
		if tokens[i].kind == tokenRBrace && tokens[i].depth == openDepth+1 {
			return open, i, nil
		}
	}
	return 0, 0, errors.New("naiveruntime: unterminated forward_proxy block")
}

func parseCredentialLines(input []byte, tokens []token, openIndex, closeIndex int) ([]credentialLine, error) {
	directDepth := tokens[openIndex].depth + 1
	var lines []credentialLine

	for i := openIndex + 1; i < closeIndex; i++ {
		tok := tokens[i]
		if tok.kind != tokenWord || tok.value != "basic_auth" {
			continue
		}
		if tok.depth != directDepth || !tok.firstOnLine {
			return nil, errors.New("naiveruntime: ambiguous nested or inline basic_auth directive")
		}

		lineTokens := []token{tok}
		j := i + 1
		for ; j < closeIndex && tokens[j].line == tok.line; j++ {
			lineTokens = append(lineTokens, tokens[j])
		}
		if len(lineTokens) != 3 || !isValueToken(lineTokens[1]) || !isValueToken(lineTokens[2]) {
			return nil, errors.New("naiveruntime: basic_auth must contain exactly username and password")
		}
		if lineTokens[1].depth != directDepth || lineTokens[2].depth != directDepth {
			return nil, errors.New("naiveruntime: basic_auth arguments have ambiguous brace depth")
		}

		username, err := tokenValue(input, lineTokens[1])
		if err != nil {
			return nil, fmt.Errorf("naiveruntime: parse basic_auth username: %w", err)
		}
		password, err := tokenValue(input, lineTokens[2])
		if err != nil {
			return nil, fmt.Errorf("naiveruntime: parse basic_auth password: %w", err)
		}
		if err := runtimecred.ValidateUsername(username); err != nil {
			return nil, fmt.Errorf("naiveruntime: unsafe imported username: %w", err)
		}
		if err := runtimecred.ValidatePassword(password, true); err != nil {
			return nil, fmt.Errorf("naiveruntime: unsafe imported password: %w", err)
		}

		lineStart, lineEnd, newline := lineBounds(input, tok.start)
		if lineStart < 0 || lineEnd < lineTokens[2].end {
			return nil, errors.New("naiveruntime: invalid basic_auth line bounds")
		}
		prefix := input[lineStart:tok.start]
		if len(bytes.Trim(prefix, " \t")) != 0 {
			return nil, errors.New("naiveruntime: basic_auth line has non-whitespace prefix")
		}
		suffixEnd := lineEnd - len(newline)
		if suffixEnd < lineTokens[2].end || len(bytes.TrimSpace(input[lineTokens[2].end:suffixEnd])) != 0 {
			return nil, errors.New("naiveruntime: basic_auth line has ambiguous trailing content")
		}

		lines = append(lines, credentialLine{
			username: username,
			password: password,
			start:    lineStart,
			end:      lineEnd,
			indent:   string(prefix),
			newline:  newline,
		})
		i = j - 1
	}
	return lines, nil
}

func lexCaddyfile(input []byte) ([]token, error) {
	tokens := make([]token, 0, 64)
	depth := 0
	line := 1
	haveTokenOnLine := false

	for i := 0; i < len(input); {
		switch input[i] {
		case ' ', '\t', '\r':
			i++
			continue
		case '\n':
			line++
			haveTokenOnLine = false
			i++
			continue
		case '#':
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		case '{':
			tokens = append(tokens, token{kind: tokenLBrace, start: i, end: i + 1, line: line, depth: depth, firstOnLine: !haveTokenOnLine})
			haveTokenOnLine = true
			depth++
			i++
			continue
		case '}':
			if depth == 0 {
				return nil, errors.New("naiveruntime: unmatched closing brace")
			}
			tokens = append(tokens, token{kind: tokenRBrace, start: i, end: i + 1, line: line, depth: depth, firstOnLine: !haveTokenOnLine})
			haveTokenOnLine = true
			depth--
			i++
			continue
		case '"':
			start := i
			first := !haveTokenOnLine
			i++
			escaped := false
			for i < len(input) {
				b := input[i]
				if b == '\n' || b == '\r' {
					return nil, errors.New("naiveruntime: newline in quoted token")
				}
				if escaped {
					escaped = false
					i++
					continue
				}
				if b == '\\' {
					escaped = true
					i++
					continue
				}
				if b == '"' {
					i++
					tokens = append(tokens, token{kind: tokenQuoted, start: start, end: i, line: line, depth: depth, firstOnLine: first})
					haveTokenOnLine = true
					goto next
				}
				i++
			}
			return nil, errors.New("naiveruntime: unterminated quoted token")
		default:
			start := i
			first := !haveTokenOnLine
			for i < len(input) {
				b := input[i]
				if b == ' ' || b == '\t' || b == '\r' || b == '\n' || b == '{' || b == '}' || b == '#' || b == '"' {
					break
				}
				i++
			}
			if i == start {
				return nil, fmt.Errorf("naiveruntime: unsupported byte 0x%02x", input[i])
			}
			tokens = append(tokens, token{kind: tokenWord, value: string(input[start:i]), start: start, end: i, line: line, depth: depth, firstOnLine: first})
			haveTokenOnLine = true
		}
	next:
	}
	if depth != 0 {
		return nil, errors.New("naiveruntime: unmatched opening brace")
	}
	return tokens, nil
}

func isValueToken(tok token) bool {
	return tok.kind == tokenWord || tok.kind == tokenQuoted
}

func tokenValue(input []byte, tok token) (string, error) {
	if tok.kind == tokenWord {
		return string(input[tok.start:tok.end]), nil
	}
	if tok.kind != tokenQuoted || tok.end-tok.start < 2 {
		return "", errors.New("not a value token")
	}

	raw := input[tok.start+1 : tok.end-1]
	var out strings.Builder
	out.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] == '\\' && i+1 < len(raw) && raw[i+1] == '"' {
			out.WriteByte('"')
			i++
			continue
		}
		out.WriteByte(raw[i])
	}
	return out.String(), nil
}

func quoteCaddyToken(value string) (string, error) {
	if strings.HasSuffix(value, "\\") {
		return "", errors.New("password ending in backslash is not safely representable by the conservative Caddy renderer")
	}
	for i := 1; i < len(value); i++ {
		if value[i] == '"' && value[i-1] == '\\' {
			return "", errors.New("password containing backslash immediately before quote is not safely representable by the conservative Caddy renderer")
		}
	}

	var out strings.Builder
	out.Grow(len(value) + 2)
	out.WriteByte('"')
	for i := 0; i < len(value); i++ {
		if value[i] == '"' {
			out.WriteByte('\\')
		}
		out.WriteByte(value[i])
	}
	out.WriteByte('"')
	return out.String(), nil
}

func lineBounds(input []byte, offset int) (start, end int, newline string) {
	start = bytes.LastIndexByte(input[:offset], '\n') + 1
	if nl := bytes.IndexByte(input[offset:], '\n'); nl >= 0 {
		end = offset + nl + 1
		if end >= 2 && input[end-2] == '\r' {
			newline = "\r\n"
		} else {
			newline = "\n"
		}
		return start, end, newline
	}
	return start, len(input), ""
}
