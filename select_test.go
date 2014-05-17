package dbr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectBasicToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").From("c").Where("id = ?", 1).ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c WHERE (id = ?)")
	assert.Equal(t, args, []interface{}{1})
}

func TestSelectFullToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").
		Distinct().
		From("c").
		Where("d = ? OR e = ?", 1, "wat").
		// Where(Eq{"f": 2}).
		// Where(map[string]interface{}{"g": 3}).
		// Where(Eq{"h": []int{4, 5, 6}}).
		GroupBy("e").
		Having("f = g").
		OrderBy("h").
		Limit(7).
		Offset(8).ToSql()

	assert.Equal(t, sql, "SELECT DISTINCT a, b FROM c WHERE (d = ? OR e = ?) GROUP BY e HAVING (f = g) ORDER BY h LIMIT 7 OFFSET 8")
	assert.Equal(t, args, []interface{}{1, "wat"})
}

func TestSelectPaginateOrderDirToSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").
		From("c").
		Where("d = ?", 1).
		Paginate(1, 20).
		OrderDir("id", false).
		ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c WHERE (d = ?) ORDER BY id DESC LIMIT 20 OFFSET 0")
	assert.Equal(t, args, []interface{}{1})

	sql, args = s.Select("a", "b").
		From("c").
		Where("d = ?", 1).
		Paginate(3, 30).
		OrderDir("id", true).
		ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c WHERE (d = ?) ORDER BY id ASC LIMIT 30 OFFSET 60")
	assert.Equal(t, args, []interface{}{1})
}

func TestSelectNoWhereSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").From("c").ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c")
	assert.Equal(t, args, []interface{}(nil))
}

func TestSelectMultiHavingSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").From("c").Where("p = ?", 1).GroupBy("z").Having("z = ?", 2).Having("y = ?", 3).ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c WHERE (p = ?) GROUP BY z HAVING (z = ?) AND (y = ?)")
	assert.Equal(t, args, []interface{}{1, 2, 3})
}

func TestSelectMultiOrderSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a", "b").From("c").OrderBy("name ASC").OrderBy("id DESC").ToSql()

	assert.Equal(t, sql, "SELECT a, b FROM c ORDER BY name ASC, id DESC")
	assert.Equal(t, args, []interface{}(nil))
}

func TestSelectWhereMapSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a").From("b").Where(map[string]interface{}{"a": 1}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a = ?)")
	assert.Equal(t, args, []interface{}{1})

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": 1, "b": true}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE ((a = ?) AND (b = ?))")
	assert.Equal(t, args, []interface{}{1, true})

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": nil}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a IS NULL)")
	assert.Equal(t, args, []interface{}(nil))

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{1, 2, 3}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a IN ?)")
	assert.Equal(t, args, []interface{}{[]int{1, 2, 3}})

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{1}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a = ?)")
	assert.Equal(t, args, []interface{}{1})

	// NOTE: a has no valid values, we want a query that returns nothing
	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (1=0)")
	assert.Equal(t, args, []interface{}(nil))

	var aval []int
	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": aval}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a IS NULL)")
	assert.Equal(t, args, []interface{}(nil))

	sql, args = s.Select("a").From("b").
		Where(map[string]interface{}{"a": []int(nil)}).
		Where(map[string]interface{}{"b": false}).
		ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (a IS NULL) AND (b = ?)")
	assert.Equal(t, args, []interface{}{false})
}

func TestSelectWhereEqSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a").From("b").Where(Eq{"a": 1, "b": []int64{1, 2, 3}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE ((a = ?) AND (b IN ?))")
	assert.Equal(t, args, []interface{}{1, []int64{1, 2, 3}})
}

func TestSelectBySql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.SelectBySql("SELECT * FROM users WHERE x = 1").ToSql()
	assert.Equal(t, sql, "SELECT * FROM users WHERE x = 1")
	assert.Equal(t, args, []interface{}(nil))

	sql, args = s.SelectBySql("SELECT * FROM users WHERE x = ? AND y IN ?", 9, []int{5, 6, 7}).ToSql()
	assert.Equal(t, sql, "SELECT * FROM users WHERE x = ? AND y IN ?")
	assert.Equal(t, args, []interface{}{9, []int{5, 6, 7}})

	// Doesn't fix shit if it's broken:
	sql, args = s.SelectBySql("wat", 9, []int{5, 6, 7}).ToSql()
	assert.Equal(t, sql, "wat")
	assert.Equal(t, args, []interface{}{9, []int{5, 6, 7}})
}

func TestSelectVarieties(t *testing.T) {
	s := createFakeSession()

	sql, _ := s.Select("id, name, email").From("users").ToSql()
	sql2, _ := s.Select("id", "name", "email").From("users").ToSql()
	assert.Equal(t, sql, sql2)
}

func TestSelectLoadAll(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.Select("id", "name", "email").From("dbr_people").OrderBy("id ASC").LoadAll(&people)

	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	assert.Equal(t, len(people), 2)
	if len(people) == 2 {
		// Make sure that the Ids are set. It's possible (maybe?) that different DBs set ids differently so
		// don't assume they're 1 and 2.
		assert.True(t, people[0].Id > 0)
		assert.True(t, people[1].Id > people[0].Id)

		assert.Equal(t, people[0].Name, "Jonathan")
		assert.True(t, people[0].Email.Valid)
		assert.Equal(t, people[0].Email.String, "jonathan@uservoice.com")
		assert.Equal(t, people[1].Name, "Dmitri")
		assert.True(t, people[1].Email.Valid)
		assert.Equal(t, people[1].Email.String, "zavorotni@jadius.com")
	}

	// TODO: test map
}

func TestSelectLoadOne(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Found:
	var person dbrPerson
	err := s.Select("id", "name", "email").From("dbr_people").Where("email = ?", "jonathan@uservoice.com").LoadOne(&person)
	assert.NoError(t, err)
	assert.True(t, person.Id > 0)
	assert.Equal(t, person.Name, "Jonathan")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "jonathan@uservoice.com")

	// Not found:
	var person2 dbrPerson
	err = s.Select("id", "name", "email").From("dbr_people").Where("email = ?", "dontexist@uservoice.com").LoadOne(&person2)
	assert.Equal(t, err, ErrNotFound)
}

func TestSelectBySqlLoadAll(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.SelectBySql("SELECT name FROM dbr_people WHERE email IN ?", []string{"jonathan@uservoice.com"}).LoadAll(&people)

	assert.NoError(t, err)
	assert.Equal(t, count, 1)
	if len(people) == 1 {
		assert.Equal(t, people[0].Name, "Jonathan")
		assert.Equal(t, people[0].Id, 0)              // not set
		assert.Equal(t, people[0].Email.Valid, false) // not set
		assert.Equal(t, people[0].Email.String, "")   // not set
	}
}

// Series of tests that test mapping struct fields to columns
