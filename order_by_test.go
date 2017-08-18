package sqlb

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestOrderByClauseSingleAsc(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cd := &ColumnDef{
        name: "name",
        table: td,
    }

    ob := &OrderByClause{
        cols: []*sortColumn{
            &sortColumn{el: cd},
        },
    }

    exp := " ORDER BY name"
    expLen := len(exp)
    expArgCount := 0

    s := ob.Size()
    assert.Equal(expLen, s)

    argc := ob.ArgCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := ob.Scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestOrderByClauseSingleDesc(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cd := &ColumnDef{
        name: "name",
        table: td,
    }

    ob := &OrderByClause{
        cols: []*sortColumn{
            &sortColumn{el: cd, desc: true},
        },
    }

    exp := " ORDER BY name DESC"
    expLen := len(exp)
    expArgCount := 0

    s := ob.Size()
    assert.Equal(expLen, s)

    argc := ob.ArgCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := ob.Scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestOrderByClauseMultiAsc(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cd1 := &ColumnDef{
        name: "name",
        table: td,
    }

    cd2 := &ColumnDef{
        name: "email",
        table: td,
    }

    ob := &OrderByClause{
        cols: []*sortColumn{
            &sortColumn{el: cd1},
            &sortColumn{el: cd2},
        },
    }

    exp := " ORDER BY name, email"
    expLen := len(exp)
    expArgCount := 0

    s := ob.Size()
    assert.Equal(expLen, s)

    argc := ob.ArgCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := ob.Scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestOrderByClauseMultiAscDesc(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cd1 := &ColumnDef{
        name: "name",
        table: td,
    }

    cd2 := &ColumnDef{
        name: "email",
        table: td,
    }

    ob := &OrderByClause{
        cols: []*sortColumn{
            &sortColumn{el: cd1},
            &sortColumn{el: cd2, desc: true},
        },
    }

    exp := " ORDER BY name, email DESC"
    expLen := len(exp)
    expArgCount := 0

    s := ob.Size()
    assert.Equal(expLen, s)

    argc := ob.ArgCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := ob.Scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}
