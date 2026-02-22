package csmith

import (
	"fmt"
	"strings"
)

// Generate emits a minimal valid C program. This is the initial foundation for the Csmith port.
func Generate(opts Options) (string, error) {
	var b strings.Builder
	b.WriteString("/* csmith-go: seed = ")
	b.WriteString(fmt.Sprintf("%d", opts.Seed))
	b.WriteString(" */\n")
	b.WriteString("#include <stdint.h>\n")
	b.WriteString("\n")
	b.WriteString("int main(void) {\n")
	b.WriteString("    return 0;\n")
	b.WriteString("}\n")
	return b.String(), nil
}
