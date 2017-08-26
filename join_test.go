package sqlb

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

var (
    meta = &Meta{
        schemaName: "test",
        tdefs: make(map[string]*TableDef, 0),
    }

    users = &TableDef{
        meta: meta,
        name: "users",
    }
    colUserId = &ColumnDef{
        name: "id",
        tdef: users,
    }
    colUserName = &ColumnDef{
        name: "name",
        tdef: users,
    }

    articles = &TableDef{
        meta: meta,
        name: "articles",
    }
    colArticleId = &ColumnDef{
        name: "id",
        tdef: articles,
    }
    colArticleAuthor = &ColumnDef{
        name: "author",
        tdef: articles,
    }
)

func init() {
    users.cdefs = []*ColumnDef{colUserId, colUserName}
    articles.cdefs = []*ColumnDef{colArticleId, colArticleAuthor}
    meta.tdefs["users"] = users
    meta.tdefs["articles"] = articles
}

func TestJoinFuncGenerics(t *testing.T) {
    // Test that the sqlb.Join() func can take a *Table or *TableDef and zero
    // or more *Expression struct pointers and returns a *joinClause struct
    // pointers. Essentially, we're testing the Selection generic interface here
    assert := assert.New(t)

    cond := Equal(colArticleAuthor, colUserId)

    joins := []*joinClause{
        Join(articles, users, cond),
        Join(articles.Table(), users.Table(), cond),
    }

    for _, j := range joins {
        exp := " JOIN users ON articles.author = users.id"
        expLen := len(exp)
        expArgCount := 0

        s := j.size()
        assert.Equal(expLen, s)

        argc := j.argCount()
        assert.Equal(expArgCount, argc)

        args := make([]interface{}, expArgCount)
        b := make([]byte, s)
        written, numArgs := j.scan(b, args)

        assert.Equal(s, written)
        assert.Equal(exp, string(b))
        assert.Equal(expArgCount, numArgs)
    }
}

func TestjoinClauseInnerOnEqualSingle(t *testing.T) {
    assert := assert.New(t)

    j := &joinClause{
        left: articles.Table(),
        right: users.Table(),
        onExprs: []*Expression{
            Equal(colArticleAuthor, colUserId),
        },
    }

    exp := " JOIN users ON articles.author = users.id"
    expLen := len(exp)
    expArgCount := 0

    s := j.size()
    assert.Equal(expLen, s)

    argc := j.argCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := j.scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestjoinClauseOnMethod(t *testing.T) {
    assert := assert.New(t)

    j := &joinClause{
        left: articles.Table(),
        right: users.Table(),
    }
    j.On(Equal(colArticleAuthor, colUserId))

    exp := " JOIN users ON articles.author = users.id"
    expLen := len(exp)
    expArgCount := 0

    s := j.size()
    assert.Equal(expLen, s)

    argc := j.argCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := j.scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestjoinClauseAliasedInnerOnEqualSingle(t *testing.T) {
    assert := assert.New(t)

    atbl := articles.Table().As("a")
    utbl := users.Table().As("u")

    aliasAuthorCol := atbl.Column("author")
    assert.NotNil(aliasAuthorCol)

    aliasIdCol := utbl.Column("id")
    assert.NotNil(aliasIdCol)

    j := &joinClause{
        left: atbl,
        right: utbl,
        onExprs: []*Expression{
            Equal(aliasAuthorCol, aliasIdCol),
        },
    }

    exp := " JOIN users AS u ON a.author = u.id"
    expLen := len(exp)
    expArgCount := 0

    s := j.size()
    assert.Equal(expLen, s)

    argc := j.argCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := j.scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
}

func TestjoinClauseInnerOnEqualMulti(t *testing.T) {
    assert := assert.New(t)

    j := &joinClause{
        left: articles.Table(),
        right: users.Table(),
        onExprs: []*Expression{
            Equal(colArticleAuthor, colUserId),
            Equal(colUserName, "foo"),
        },
    }

    exp := " JOIN users ON articles.author = users.id AND users.name = ?"
    expLen := len(exp)
    expArgCount := 1

    s := j.size()
    assert.Equal(expLen, s)

    argc := j.argCount()
    assert.Equal(expArgCount, argc)

    args := make([]interface{}, expArgCount)
    b := make([]byte, s)
    written, numArgs := j.scan(b, args)

    assert.Equal(s, written)
    assert.Equal(exp, string(b))
    assert.Equal(expArgCount, numArgs)
    assert.Equal("foo", args[0])
}
