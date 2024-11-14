//go:generate go run github.com/dmarkham/enumer -json -trimprefix Type -type StdioType
package runner

type StdioType int

const (
	TypeStdout StdioType = iota
	TypeStderr
	TypeStdin
)
