package dbr

import (
	// "database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransactionReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	tx, err := s.Begin()
	assert.NoError(t, err)

	res, err := tx.InsertInto("dbr_people").Columns("name", "email").Values("Barack", "obama@whitehouse.gov").Exec()

	assert.NoError(t, err)
	id, err := res.LastInsertId()
	assert.NoError(t, err)
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)

	assert.True(t, id > 0)
	assert.Equal(t, rowsAff, 1)

	var person dbrPerson
	err = tx.Select("*").From("dbr_people").Where("id = ?", id).LoadOne(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Id, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "obama@whitehouse.gov")

	err = tx.Commit()
	assert.NoError(t, err)
}

func TestTransactionRollbackReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()

	tx, err := s.Begin()
	assert.NoError(t, err)

	var person dbrPerson
	err = tx.Select("*").From("dbr_people").Where("email = ?", "jonathan@uservoice.com").LoadOne(&person)
	assert.NoError(t, err)
	assert.Equal(t, person.Name, "Jonathan")

	err = tx.Rollback()
	assert.NoError(t, err)
}
