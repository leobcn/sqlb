package sqlb

type groupByClause struct {
	cols []projection
}

func (gb *groupByClause) argCount() int {
	argc := 0
	return argc
}

func (gb *groupByClause) size() int {
	size := len(Symbols[SYM_GROUP_BY])
	ncols := len(gb.cols)
	for _, c := range gb.cols {
		reset := c.disableAliasScan()
		defer reset()
		size += c.size()
	}
	return size + (len(Symbols[SYM_COMMA_WS]) * (ncols - 1)) // the commas...
}

func (gb *groupByClause) scan(b []byte, args []interface{}, curArg *int) int {
	bw := 0
	bw += copy(b[bw:], Symbols[SYM_GROUP_BY])
	ncols := len(gb.cols)
	for x, c := range gb.cols {
		reset := c.disableAliasScan()
		defer reset()
		bw += c.scan(b[bw:], args, curArg)
		if x != (ncols - 1) {
			bw += copy(b[bw:], Symbols[SYM_COMMA_WS])
		}
	}
	return bw
}
