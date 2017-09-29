package sqlb

import (
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestSelectQuery(t *testing.T) {
    assert := assert.New(t)

    m := testFixtureMeta()
    users := m.Table("users")
    articles := m.Table("articles")
    articleStates := m.Table("article_states")
    userProfiles := m.Table("user_profiles")
    colUserId := users.Column("id")
    colUserName := users.Column("name")
    colArticleId := articles.Column("id")
    colArticleAuthor := articles.Column("author")
    colArticleState := articles.Column("state")
    colArticleStateId := articleStates.Column("id")
    colArticleStateName := articleStates.Column("name")
    colUserProfileContent := userProfiles.Column("content")
    colUserProfileUser := userProfiles.Column("user")

    subq := Select(colUserId).As("users_derived")

    tests := []struct{
        name string
        q *SelectQuery
        qs string
        qargs []interface{}
        qe error
    }{
        {
            name: "Simple FROM",
            q: Select(users),
            qs: "SELECT users.id, users.name FROM users",
        },
        {
            name: "Simple SELECT COUNT(*) FROM",
            q: Select(Count(users)),
            qs: "SELECT COUNT(*) FROM users",
        },
        {
            name: "Simple WHERE",
            q: Select(users).Where(Equal(colUserName, "foo")),
            qs: "SELECT users.id, users.name FROM users WHERE users.name = ?",
            qargs: []interface{}{"foo"},
        },
        {
            name: "Simple GROUP BY",
            q: Select(users).GroupBy(colUserName),
            qs: "SELECT users.id, users.name FROM users GROUP BY users.name",
        },
        {
            name: "Simple ORDER BY",
            q: Select(users).OrderBy(colUserName.Desc()),
            qs: "SELECT users.id, users.name FROM users ORDER BY users.name DESC",
        },
        {
            name: "Simple LIMIT",
            q: Select(users).Limit(10),
            qs: "SELECT users.id, users.name FROM users LIMIT ?",
            qargs: []interface{}{10},
        },
        {
            name: "Simple LIMIT with OFFSET",
            q: Select(users).LimitWithOffset(10, 20),
            qs: "SELECT users.id, users.name FROM users LIMIT ? OFFSET ?",
            qargs: []interface{}{10, 20},
        },
        {
            name: "Simple named derived table",
            q: Select(Select(users).As("u")),
            qs: "SELECT u.id, u.name FROM (SELECT users.id, users.name FROM users) AS u",
        },
        {
            name: "Simple un-named derived table",
            q: Select(Select(users)),
            qs: "SELECT derived0.id, derived0.name FROM (SELECT users.id, users.name FROM users) AS derived0",
        },
        {
            name: "Bad JOIN. Can't Join() against no selection",
            q: Select().Join(users, Equal(colArticleAuthor, colUserId)),
            qe: ERR_JOIN_INVALID_NO_SELECT,
        },
        {
            name: "Bad JOIN. Can't Join() against a selection that isn't in the containing SELECT",
            q: Select(articleStates).Join(users, Equal(colArticleAuthor, colUserId)),
            qe: ERR_JOIN_INVALID_UNKNOWN_TARGET,
        },
        {
            name: "Simple INNER JOIN",
            q: Select(colArticleId, colUserName.As("author")).Join(users, Equal(colArticleAuthor, colUserId)),
            qs: "SELECT articles.id, users.name AS author FROM articles JOIN users ON articles.author = users.id",
        },
        {
            name: "Simple LEFT JOIN",
            q: Select(colArticleId, colUserName.As("author")).OuterJoin(users, Equal(colArticleAuthor, colUserId)),
            qs: "SELECT articles.id, users.name AS author FROM articles LEFT JOIN users ON articles.author = users.id",
        },
        {
            name: "JOIN A to B and A to C",
            q: Select(
                colArticleId,
                colUserName.As("author"),
                colArticleStateName.As("state"),
            ).Join(users, Equal(colArticleAuthor, colUserId),
            ).Join(articleStates, Equal(colArticleState, colArticleStateId)),
            qs: "SELECT articles.id, users.name AS author, article_states.name AS state FROM articles JOIN users ON articles.author = users.id JOIN article_states ON articles.state = article_states.id",
        },
        {
            name: "LEFT JOIN to derived table (subquery in FROM clause)",
            q: Select(
                colUserId,
                colUserName,
            ).OuterJoin(subq, Equal(colUserId, subq.Column("id"))),
            qs: "SELECT users.id, users.name FROM users LEFT JOIN (SELECT users.id FROM users) AS users_derived ON users.id = users_derived.id",
        },
        {
            name: "JOIN to derived table (subquery in FROM clause)",
            q: Select(
                colUserId,
                colUserName,
            ).Join(subq, Equal(colUserId, subq.Column("id"))),
            qs: "SELECT users.id, users.name FROM users JOIN (SELECT users.id FROM users) AS users_derived ON users.id = users_derived.id",
        },
        {
            name: "JOIN A to B and B to C",
            q: Select(
                colArticleId,
                colUserName.As("author"),
                colUserProfileContent.As("author_profile"),
            ).Join(users, Equal(colArticleAuthor, colUserId),
            ).Join(userProfiles, Equal(colUserId, colUserProfileUser)),
            qs: "SELECT articles.id, users.name AS author, user_profiles.content AS author_profile FROM articles JOIN users ON articles.author = users.id JOIN user_profiles ON users.id = user_profiles.user",
        },
    }
    for _, test := range tests {
        if test.qe != nil {
            assert.Equal(test.qe, test.q.Error())
            continue
        } else if test.q.Error() != nil {
            qe := test.q.Error()
            assert.Fail(qe.Error())
            continue
        }
        qs, qargs := test.q.StringArgs()
        assert.Equal(len(test.qargs), len(qargs))
        assert.Equal(test.qs, qs)
    }
}

func TestNestedSetQueries(t *testing.T) {
    // ref: https://github.com/jaypipes/sqlb/issues/49
    assert := assert.New(t)

    m := &Meta{}
    orgs := &TableDef{
        meta: m,
        name: "organizations",
    }
    cols := []*ColumnDef{
        &ColumnDef{
            tdef: orgs,
            name: "id",
        },
        &ColumnDef{
            tdef: orgs,
            name: "root_organization_id",
        },
        &ColumnDef{
            tdef: orgs,
            name: "nested_set_left",
        },
        &ColumnDef{
            tdef: orgs,
            name: "nested_set_right",
        },
    }
    orgs.cdefs = cols

    o1 := orgs.As("o1")
    o2 := orgs.As("o2")

    o1id := o1.Column("id")
    o2id := o2.Column("id")
    o1rootid := o1.Column("root_organization_id")
    o2rootid := o2.Column("root_organization_id")
    o1nestedleft := o1.Column("nested_set_left")
    o2nestedleft := o2.Column("nested_set_left")
    o2nestedright := o2.Column("nested_set_right")

    joinCond := And(
        Equal(o1rootid, o2rootid),
        Between(o1nestedleft, o2nestedleft, o2nestedright),
    )
    q := Select(o1id).Join(o2, joinCond)
    q.Where(Equal(o2id, 2))

    qs, qargs := q.StringArgs()

    expqs := "SELECT o1.id FROM organizations AS o1 JOIN organizations AS o2 ON (o1.root_organization_id = o2.root_organization_id AND o1.nested_set_left BETWEEN o2.nested_set_left AND o2.nested_set_right) WHERE o2.id = ?"
    expqargs := []interface{}{2}

    assert.Equal(expqs, qs)
    assert.Equal(expqargs, qargs)
}

func TestNestedSetWithAdditionalJoin(t *testing.T) {
    // ref: https://github.com/jaypipes/sqlb/issues/60
    assert := assert.New(t)

    m := &Meta{}
    orgs := &TableDef{
        meta: m,
        name: "organizations",
    }
    orgCols := []*ColumnDef{
        &ColumnDef{
            tdef: orgs,
            name: "id",
        },
        &ColumnDef{
            tdef: orgs,
            name: "root_organization_id",
        },
        &ColumnDef{
            tdef: orgs,
            name: "nested_set_left",
        },
        &ColumnDef{
            tdef: orgs,
            name: "nested_set_right",
        },
    }
    orgs.cdefs = orgCols

    orgUsers := &TableDef{
        meta: m,
        name: "organization_users",
    }
    orgUsersCols := []*ColumnDef{
        &ColumnDef{
            tdef: orgUsers,
            name: "organization_id",
        },
        &ColumnDef{
            tdef: orgUsers,
            name: "user_id",
        },
    }
    orgUsers.cdefs = orgUsersCols

    o1 := orgs.As("o1")
    o2 := orgs.As("o2")
    ou := orgUsers.As("ou")

    o1id := o1.Column("id")
    o2id := o2.Column("id")
    o1rootid := o1.Column("root_organization_id")
    o2rootid := o2.Column("root_organization_id")
    o1nestedleft := o1.Column("nested_set_left")
    o2nestedleft := o2.Column("nested_set_left")
    o2nestedright := o2.Column("nested_set_right")
    ouUserId := ou.Column("user_id")
    ouOrgId := ou.Column("organization_id")

    nestedJoinCond := And(
        Equal(o1rootid, o2rootid),
        Between(o1nestedleft, o2nestedleft, o2nestedright),
    )
    ouJoin := And(
        Equal(o2id, ouOrgId),
        Equal(ouUserId, 1),
    )
    q := Select(o1id).Join(o2, nestedJoinCond).Join(ou, ouJoin)

    assert.Nil(q.e)

    qs, qargs := q.StringArgs()

    expqs := "SELECT o1.id FROM organizations AS o1 JOIN organizations AS o2 ON (o1.root_organization_id = o2.root_organization_id AND o1.nested_set_left BETWEEN o2.nested_set_left AND o2.nested_set_right) JOIN organization_users AS ou ON (o2.id = ou.organization_id AND ou.user_id = ?)"
    expqargs := []interface{}{1}

    assert.Equal(expqs, qs)
    assert.Equal(expqargs, qargs)
}

func TestModifyingSelectQueryUpdatesBuffer(t *testing.T) {
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

func TestSelectQueryErrors(t *testing.T) {
    assert := assert.New(t)

    q := &SelectQuery{}

    assert.False(q.IsValid()) // Doesn't have a selectClause yet...
    assert.Nil(q.Error()) // But there is no error set yet...

    m := testFixtureMeta()
    users := m.TableDef("users")

    q = Select(users)

    assert.True(q.IsValid())
    assert.Nil(q.Error())

    q.e = fmt.Errorf("Cannot determine left side of JOIN expression.")
    assert.False(q.IsValid())
    assert.NotNil(q.Error())
}
