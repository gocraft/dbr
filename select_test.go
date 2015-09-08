package dbr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkSelectBasicSql(b *testing.B) {
	s := createFakeSession()

	// Do some allocations outside the loop so they don't affect the results
	argEq := Eq{"a": []int{1, 2, 3}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Select("something_id", "user_id", "other").
			From("some_table").
			Where("d = ? OR e = ?", 1, "wat").
			Where(argEq).
			OrderDir("id", false).
			Paginate(1, 20).
			ToSql()
	}
}

func BenchmarkSelectFullSql(b *testing.B) {
	s := createFakeSession()

	// Do some allocations outside the loop so they don't affect the results
	argEq1 := Eq{"f": 2, "x": "hi"}
	argEq2 := map[string]interface{}{"g": 3}
	argEq3 := Eq{"h": []int{1, 2, 3}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Select("a", "b", "z", "y", "x").
			Distinct().
			From("c").
			Where("d = ? OR e = ?", 1, "wat").
			Where(argEq1).
			Where(argEq2).
			Where(argEq3).
			GroupBy("i").
			GroupBy("ii").
			GroupBy("iii").
			Having("j = k").
			Having("jj = ?", 1).
			Having("jjj = ?", 2).
			OrderBy("l").
			OrderBy("l").
			OrderBy("l").
			Limit(7).
			Offset(8).
			ToSql()
	}
}

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
		Where(Eq{"f": 2}).
		Where(map[string]interface{}{"g": 3}).
		Where(Eq{"h": []int{4, 5, 6}}).
		GroupBy("i").
		Having("j = k").
		OrderBy("l").
		Limit(7).
		Offset(8).
		ToSql()

	assert.Equal(t, sql, "SELECT DISTINCT a, b FROM c WHERE (d = ? OR e = ?) AND (`f` = ?) AND (`g` = ?) AND (`h` IN ?) GROUP BY i HAVING (j = k) ORDER BY l LIMIT 7 OFFSET 8")
	assert.Equal(t, args, []interface{}{1, "wat", 2, 3, []int{4, 5, 6}})
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
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` = ?)")
	assert.Equal(t, args, []interface{}{1})

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": 1, "b": true}).ToSql()
	if sql == "SELECT a FROM b WHERE (`a` = ?) AND (`b` = ?)" {
		assert.Equal(t, args, []interface{}{1, true})
	} else {
		assert.Equal(t, sql, "SELECT a FROM b WHERE (`b` = ?) AND (`a` = ?)")
		assert.Equal(t, args, []interface{}{true, 1})
	}

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": nil}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` IS NULL)")
	assert.Equal(t, args, []interface{}(nil))

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{1, 2, 3}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` IN ?)")
	assert.Equal(t, args, []interface{}{[]int{1, 2, 3}})

	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{1}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` = ?)")
	assert.Equal(t, args, []interface{}{1})

	// NOTE: a has no valid values, we want a query that returns nothing
	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": []int{}}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (1=0)")
	assert.Equal(t, args, []interface{}(nil))

	var aval []int
	sql, args = s.Select("a").From("b").Where(map[string]interface{}{"a": aval}).ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` IS NULL)")
	assert.Equal(t, args, []interface{}(nil))

	sql, args = s.Select("a").From("b").
		Where(map[string]interface{}{"a": []int(nil)}).
		Where(map[string]interface{}{"b": false}).
		ToSql()
	assert.Equal(t, sql, "SELECT a FROM b WHERE (`a` IS NULL) AND (`b` = ?)")
	assert.Equal(t, args, []interface{}{false})
}

func TestSelectWhereEqSql(t *testing.T) {
	s := createFakeSession()

	sql, args := s.Select("a").From("b").Where(Eq{"a": 1, "b": []int64{1, 2, 3}}).ToSql()
	if sql == "SELECT a FROM b WHERE (`a` = ?) AND (`b` IN ?)" {
		assert.Equal(t, args, []interface{}{1, []int64{1, 2, 3}})
	} else {
		assert.Equal(t, sql, "SELECT a FROM b WHERE (`b` IN ?) AND (`a` = ?)")
		assert.Equal(t, args, []interface{}{[]int64{1, 2, 3}, 1})
	}
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

func TestSelectLoadStructs(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.Select("id", "name", "email").From("dbr_people").OrderBy("id ASC").LoadStructs(&people)

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

func TestSelectLoadStruct(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Found:
	var person dbrPerson
	err := s.Select("id", "name", "email").From("dbr_people").Where("email = ?", "jonathan@uservoice.com").LoadStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.Id > 0)
	assert.Equal(t, person.Name, "Jonathan")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "jonathan@uservoice.com")

	// Not found:
	var person2 dbrPerson
	err = s.Select("id", "name", "email").From("dbr_people").Where("email = ?", "dontexist@uservoice.com").LoadStruct(&person2)
	assert.Equal(t, err, ErrNotFound)
}

func TestSelectBySqlLoadStructs(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.SelectBySql("SELECT name FROM dbr_people WHERE email IN ?", []string{"jonathan@uservoice.com"}).LoadStructs(&people)

	assert.NoError(t, err)
	assert.Equal(t, count, 1)
	if len(people) == 1 {
		assert.Equal(t, people[0].Name, "Jonathan")
		assert.Equal(t, people[0].Id, int64(0))       // not set
		assert.Equal(t, people[0].Email.Valid, false) // not set
		assert.Equal(t, people[0].Email.String, "")   // not set
	}
}

func TestSelectLoadValue(t *testing.T) {
	s := createRealSessionWithFixtures()

	var name string
	err := s.Select("name").From("dbr_people").Where("email = 'jonathan@uservoice.com'").LoadValue(&name)

	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var id int64
	err = s.Select("id").From("dbr_people").Limit(1).LoadValue(&id)

	assert.NoError(t, err)
	assert.True(t, id > 0)
}

func TestSelectLoadValues(t *testing.T) {
	s := createRealSessionWithFixtures()

	var names []string
	count, err := s.Select("name").From("dbr_people").LoadValues(&names)

	assert.NoError(t, err)
	assert.Equal(t, count, 2)
	assert.Equal(t, names, []string{"Jonathan", "Dmitri"})

	var ids []int64
	count, err = s.Select("id").From("dbr_people").Limit(1).LoadValues(&ids)

	assert.NoError(t, err)
	assert.Equal(t, count, 1)
	assert.Equal(t, ids, []int64{1})
}

func TestSelectReturn(t *testing.T) {
	s := createRealSessionWithFixtures()

	name, err := s.Select("name").From("dbr_people").Where("email = 'jonathan@uservoice.com'").ReturnString()
	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	count, err := s.Select("COUNT(*)").From("dbr_people").ReturnInt64()
	assert.NoError(t, err)
	assert.Equal(t, count, int64(2))

	names, err := s.Select("name").From("dbr_people").Where("email = 'jonathan@uservoice.com'").ReturnStrings()
	assert.NoError(t, err)
	assert.Equal(t, names, []string{"Jonathan"})

	counts, err := s.Select("COUNT(*)").From("dbr_people").ReturnInt64s()
	assert.NoError(t, err)
	assert.Equal(t, counts, []int64{2})
}

// Series of tests that test mapping struct fields to columns
