package sqlb

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

type queryTest struct {
    q *Query
    qs string
    qargs []interface{}
}

func TestQuery(t *testing.T) {
    assert := assert.New(t)

    m := testFixtureMeta()
    users := m.Table("users")
    colUserName := users.Column("name")

    tests := []queryTest{
        // Simple FROM
        queryTest{
            q: Select(users),
            qs: "SELECT users.id, users.name FROM users",
        },
        // add WHERE
        queryTest{
            q: Select(users).Where(Equal(colUserName, "foo")),
            qs: "SELECT users.id, users.name FROM users WHERE users.name = ?",
            qargs: []interface{}{"foo"},
        },
    }
    for _, test := range tests {
        qs, qargs := test.q.StringArgs()
        assert.Equal(len(test.qargs), len(qargs))
        assert.Equal(test.qs, qs)
    }
}

func TestModifyingQueryUpdatesBuffer(t *testing.T) {
    assert := assert.New(t)

    m := testFixtureMeta()
    users := m.TableDef("users")

    q := Select(users)

    qs, qargs := q.StringArgs()
    assert.Equal("SELECT users.id, users.name FROM users", qs)
    assert.Nil(qargs)

    // Modify the underlying SELECT and verify string and args changed
    q.Where(Equal(users.Column("id"), 1))
    qs, qargs = q.StringArgs()
    assert.Equal("SELECT users.id, users.name FROM users WHERE users.id = ?", qs)
    assert.Equal([]interface{}{1}, qargs)
}
