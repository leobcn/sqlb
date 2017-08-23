package sqlb

// A Projection is something that produces a scalar value. A column, column
// definition, function, etc.
type Projection interface {
    projectionId() uint64
    // Projections must also implement Element
    Size() int
    ArgCount() int
    Scan([]byte, []interface{}) (int, int)
}

// A Selection is something that produces rows. A table, table definition,
// view, subselect, etc
type Selection interface {
    projections() []Projection
    selectionId() uint64
    // Selections must also implement Element
    Size() int
    ArgCount() int
    Scan([]byte, []interface{}) (int, int)
}

type SelectClause struct {
    alias string
    projected *List
    selections []Selection
    joins []*JoinClause
    filters []*Expression
    groupBy *GroupByClause
    orderBy *OrderByClause
    limit *LimitClause
}

func (s *SelectClause) ArgCount() int {
    argc := s.projected.ArgCount()
    for _, sel := range s.selections {
        argc += sel.ArgCount()
    }
    for _, join := range s.joins {
        argc += join.ArgCount()
    }
    for _, filter := range s.filters {
        argc += filter.ArgCount()
    }
    if s.groupBy != nil {
        argc += s.groupBy.ArgCount()
    }
    if s.orderBy != nil {
        argc += s.orderBy.ArgCount()
    }
    if s.limit != nil {
        argc += s.limit.ArgCount()
    }
    return argc
}

func (s *SelectClause) Alias(alias string) {
    s.alias = alias
}

func (s *SelectClause) As(alias string) *SelectClause {
    s.Alias(alias)
    return s
}

func (s *SelectClause) Size() int {
    size := len(Symbols[SYM_SELECT]) + len(Symbols[SYM_FROM])
    size += s.projected.Size()
    for _, sel := range s.selections {
        size += sel.Size()
    }
    if s.alias != "" {
        size += len(Symbols[SYM_AS]) + len(s.alias)
    }
    for _, join := range s.joins {
        size += join.Size()
    }
    nfilters := len(s.filters)
    if nfilters > 0 {
        size += len(Symbols[SYM_WHERE])
        size += len(Symbols[SYM_AND]) * (nfilters - 1)
        for _, filter := range s.filters {
            size += filter.Size()
        }
    }
    if s.groupBy != nil {
        size += s.groupBy.Size()
    }
    if s.orderBy != nil {
        size += s.orderBy.Size()
    }
    if s.limit != nil {
        size += s.limit.Size()
    }
    return size
}

func (s *SelectClause) Scan(b []byte, args []interface{}) (int, int) {
    var bw, ac int
    bw += copy(b[bw:], Symbols[SYM_SELECT])
    pbw, pac := s.projected.Scan(b[bw:], args)
    bw += pbw
    ac += pac
    bw += copy(b[bw:], Symbols[SYM_FROM])
    for _, sel := range s.selections {
        sbw, sac := sel.Scan(b[bw:], args)
        bw += sbw
        ac += sac
    }
    if s.alias != "" {
        bw += copy(b[bw:], Symbols[SYM_AS])
        bw += copy(b[bw:], s.alias)
    }
    for _, join := range s.joins {
        jbw, jac := join.Scan(b[bw:], args)
        bw += jbw
        ac += jac
    }
    if len(s.filters) > 0 {
        bw += copy(b[bw:], Symbols[SYM_WHERE])
        for x, filter := range s.filters {
            if x > 0 {
                bw += copy(b[bw:], Symbols[SYM_AND])
            }
            fbw, fac := filter.Scan(b[bw:], args[ac:])
            bw += fbw
            ac += fac
        }
    }
    if s.groupBy != nil {
        gbbw, gbac := s.groupBy.Scan(b[bw:], args[ac:])
        bw += gbbw
        ac += gbac
    }
    if s.orderBy != nil {
        obbw, obac := s.orderBy.Scan(b[bw:], args[ac:])
        bw += obbw
        ac += obac
    }
    if s.limit != nil {
        lbw, lac := s.limit.Scan(b[bw:], args[ac:])
        bw += lbw
        ac += lac
    }
    return bw, ac
}

func (s *SelectClause) String() string {
    size := s.Size()
    argc := s.ArgCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.Scan(b, args)
    return string(b)
}

func (s *SelectClause) StringArgs() (string, []interface{}) {
    size := s.Size()
    argc := s.ArgCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.Scan(b, args)
    return string(b), args
}

func (s *SelectClause) Where(e *Expression) *SelectClause {
    s.filters = append(s.filters, e)
    return s
}

// Given one or more columns, either set or add to the GROUP BY clause for
// the SelectClause
func (s *SelectClause) GroupBy(cols ...Columnar) *SelectClause {
    if len(cols) == 0 {
        return s
    }
    gb := s.groupBy
    if gb == nil {
        gb = &GroupByClause{
            cols: &List{
                elements: make([]Element, len(cols)),
            },
        }
        for x, c := range cols {
            gb.cols.elements[x] = c.Column()
        }
    } else {
        for _, c := range cols {
            gb.cols.elements = append(gb.cols.elements, c.Column())
        }
    }
    s.groupBy = gb
    return s
}

// Given one or more sort columns, either set or add to the ORDER BY clause for
// the SelectClause
func (s *SelectClause) OrderBy(sortCols ...*SortColumn) *SelectClause {
    if len(sortCols) == 0 {
        return s
    }
    ob := s.orderBy
    if ob == nil {
        ob = &OrderByClause{
            cols: &List{
                elements: make([]Element, len(sortCols)),
            },
        }
        for x, sc := range sortCols {
            ob.cols.elements[x] = sc
        }
    } else {
        for _, sc := range sortCols {
            ob.cols.elements = append(ob.cols.elements, sc)
        }
    }
    s.orderBy = ob
    return s
}

func (s *SelectClause) LimitWithOffset(limit int, offset int) *SelectClause {
    lc := &LimitClause{limit: limit}
    lc.offset = &offset
    s.limit = lc
    return s
}

func (s *SelectClause) Limit(limit int) *SelectClause {
    lc := &LimitClause{limit: limit}
    s.limit = lc
    return s
}

func containsJoin(s *SelectClause, j *JoinClause) bool {
    for _, sj := range s.joins {
        if j == sj {
            return true
        }
    }
    return false
}

func addToProjections(s *SelectClause, p Projection) {
    s.projected.elements = append(s.projected.elements, p)
}

func Select(items ...Element) *SelectClause {
    // TODO(jaypipes): Make the memory allocation more efficient below by
    // looping through the elements and determining the number of element struct
    // pointers to allocate instead of just making an empty array of Element
    // pointers.
    res := &SelectClause{
        projected: &List{},
    }

    selectionMap := make(map[uint64]Selection, 0)

    // For each scannable item we've received in the call, check what concrete
    // type they are and, depending on which type they are, either add them to
    // the returned SelectClause's projected List or query the underlying
    // table metadata to generate a list of all columns in that table.
    for _, item := range items {
        switch item.(type) {
            case *JoinClause:
                j := item.(*JoinClause)
                if ! containsJoin(res, j) {
                    res.joins = append(res.joins, j)
                    if _, ok := selectionMap[j.left.selectionId()]; ! ok {
                        selectionMap[j.left.selectionId()] = j.left
                        for _, proj := range j.left.projections() {
                            addToProjections(res, proj)
                        }
                    }
                }
            case *Column:
                v := item.(*Column)
                res.projected.elements = append(res.projected.elements, v)
                selectionMap[v.tbl.selectionId()] = v.tbl
            case *List:
                v := item.(*List)
                for _, el := range v.elements {
                    res.projected.elements = append(res.projected.elements, el)
                    if isColumn(el) {
                        c := el.(*Column)
                        selectionMap[c.tbl.selectionId()] = c.tbl
                    }
                }
            case *Table:
                v := item.(*Table)
                for _, cd := range v.tdef.projections() {
                    addToProjections(res, cd)
                }
                selectionMap[v.selectionId()] = v
            case *TableDef:
                v := item.(*TableDef)
                for _, cd := range v.projections() {
                    addToProjections(res, cd)
                }
                selectionMap[v.selectionId()] = v
            case *ColumnDef:
                v := item.(*ColumnDef)
                addToProjections(res, v)
                selectionMap[v.tdef.selectionId()] = v.tdef
        }
    }
    selections := make([]Selection, len(selectionMap))
    x := 0
    for _, sel := range selectionMap {
        selections[x] = sel
        x++
    }
    res.selections = selections
    return res
}
