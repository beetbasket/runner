package matcher

import (
	"bytes"
	"io"
	"slices"
)

type Matcher struct {
	parser  parse
	out     bytes.Buffer
	special bytes.Buffer
}

type Bytes interface {
	~string | ~[]byte
}

func New[B Bytes](prefix B) *Matcher {
	var m Matcher
	m.parser = newParser([]byte(prefix), &m.out, &m.special)
	return &m
}

func (m *Matcher) Write(b []byte) (n int, _ error) {
	for _, bb := range b {
		m.parser = m.parser.parse(bb)
	}
	return len(b), nil
}

func (m *Matcher) ReadSpecial() []byte {
	defer m.special.Reset()
	return slices.Clone(m.special.Bytes())
}

func (m *Matcher) ReadOut() []byte {
	defer m.out.Reset()
	return slices.Clone(m.out.Bytes())
}

type parse interface {
	parse(byte) parse
}

func newParser(prefix []byte, out, special parseWriter) parse {
	reg := &writeRegularParse{out: out}
	if len(prefix) == 0 {
		reg.head = reg
		return reg
	}

	spec := &writeSpecialParse{out: special}
	head := &prefixParse{
		spaceParse: spaceParse{
			regular: reg,
			special: spec,
			prefix:  prefix,
			out:     out,
		},
	}
	head.head = head
	reg.head = head
	spec.head = head

	curr := head
	for i := range len(prefix[1:]) {
		next := &prefixParse{
			spaceParse: spaceParse{
				head:    head,
				regular: reg,
				special: spec,
				prefix:  prefix,
				out:     out,
			},
			idx: i + 1,
		}
		curr.next = next
		curr = next
	}

	curr.next = &spaceParse{
		head:    head,
		regular: reg,
		special: spec,
		prefix:  prefix,
		out:     out,
	}

	return head
}

type spaceParse struct {
	head    parse
	regular parse
	special parse
	prefix  []byte
	out     parseWriter
}

func (sp *spaceParse) parse(b byte) parse {
	if b == ' ' {
		return sp.special
	}

	_, _ = sp.out.Write(sp.prefix)
	_ = sp.out.WriteByte(b)
	if b == '\n' {
		return sp.head
	}
	return sp.regular
}

type prefixParse struct {
	spaceParse
	next parse
	idx  int
}

func (pp *prefixParse) parse(b byte) parse {
	if b == pp.prefix[pp.idx] {
		return pp.next
	}

	_, _ = pp.out.Write(pp.prefix[:pp.idx])
	_ = pp.out.WriteByte(b)

	if b == '\n' {
		return pp.head
	}
	return pp.regular
}

type parseWriter interface {
	io.Writer
	io.ByteWriter
}

type writeSpecialParse struct {
	head parse
	buf  bytes.Buffer
	out  io.Writer
}

func (wsp *writeSpecialParse) parse(b byte) parse {
	_ = wsp.buf.WriteByte(b)
	if b == '\n' {
		_, _ = wsp.buf.WriteTo(wsp.out)
		wsp.buf.Reset()
		return wsp.head
	}
	return wsp
}

type writeRegularParse struct {
	head parse
	out  io.ByteWriter
}

func (wrp *writeRegularParse) parse(b byte) parse {
	_ = wrp.out.WriteByte(b)
	if b == '\n' {
		return wrp.head
	}
	return wrp
}
