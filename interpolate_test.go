package dbr

import (
	"strings"
	"testing"
	"time"

	"github.com/gocraft/dbr/dialect"
	"github.com/stretchr/testify/assert"
)

func TestInterpolateIgnoreBinary(t *testing.T) {
	for _, test := range []struct {
		query     string
		value     []interface{}
		wantQuery string
		wantValue []interface{}
	}{
		{
			query:     "?",
			value:     []interface{}{1},
			wantQuery: "1",
			wantValue: nil,
		},
		{
			query:     "?",
			value:     []interface{}{[]byte{1, 2, 3}},
			wantQuery: "?",
			wantValue: []interface{}{[]byte{1, 2, 3}},
		},
		{
			query:     "? ?",
			value:     []interface{}{[]byte{1}, []byte{2}},
			wantQuery: "? ?",
			wantValue: []interface{}{[]byte{1}, []byte{2}},
		},
		{
			query:     "? ?",
			value:     []interface{}{Expr("|?| ?", []byte{1}, Expr("|?|", []byte{2})), []byte{3}},
			wantQuery: "|?| |?| ?",
			wantValue: []interface{}{[]byte{1}, []byte{2}, []byte{3}},
		},
	} {
		i := interpolator{
			Buffer:       NewBuffer(),
			Dialect:      dialect.MySQL,
			IgnoreBinary: true,
		}

		err := i.interpolate(test.query, test.value)
		assert.NoError(t, err)

		assert.Equal(t, test.wantQuery, i.String())
		assert.Equal(t, test.wantValue, i.Value())
	}
}

func TestInterpolateForDialect(t *testing.T) {
	for _, test := range []struct {
		query string
		value []interface{}
		want  string
	}{
		{
			query: "?",
			value: []interface{}{nil},
			want:  "NULL",
		},
		{
			query: "?",
			value: []interface{}{`'"'"`},
			want:  "'\\'\\\"\\'\\\"'",
		},
		{
			query: "? ?",
			value: []interface{}{true, false},
			want:  "1 0",
		},
		{
			query: "? ?",
			value: []interface{}{1, 1.23},
			want:  "1 1.23",
		},
		{
			query: "?",
			value: []interface{}{time.Date(2008, 9, 17, 20, 4, 26, 123456000, time.UTC)},
			want:  "'2008-09-17 20:04:26.123456'",
		},
		{
			query: "?",
			value: []interface{}{[]string{"one", "two"}},
			want:  "('one','two')",
		},
		{
			query: "?",
			value: []interface{}{[]byte{0x1, 0x2, 0x3}},
			want:  "0x010203",
		},
		{
			query: "start?end",
			value: []interface{}{new(int)},
			want:  "start0end",
		},
		{
			query: "?",
			value: []interface{}{Select("a").From("table")},
			want:  "(SELECT a FROM table)",
		},
		{
			query: "?",
			value: []interface{}{I("a1").As("a2")},
			want:  "`a1` AS `a2`",
		},
		{
			query: "?",
			value: []interface{}{Select("a").From("table").As("a1")},
			want:  "(SELECT a FROM table) AS `a1`",
		},
		{
			query: "?",
			value: []interface{}{
				UnionAll(
					Select("a").From("table1"),
					Select("b").From("table2"),
				).As("t"),
			},
			want: "((SELECT a FROM table1) UNION ALL (SELECT b FROM table2)) AS `t`",
		},
		{
			query: "?",
			value: []interface{}{time.Month(7)},
			want:  "7",
		},
		{
			query: "?",
			value: []interface{}{(*int64)(nil)},
			want:  "NULL",
		},
	} {
		s, err := InterpolateForDialect(test.query, test.value, dialect.MySQL)
		assert.NoError(t, err)
		assert.Equal(t, test.want, s)
	}
}

// Attempts to test common SQL injection strings. See `InjectionAttempts` for
// more information on the source and the strings themselves.
func TestCommonSQLInjections(t *testing.T) {
	for _, sess := range testSession {
		for _, injectionAttempt := range strings.Split(injectionAttempts, "\n") {
			// Create a user with the attempted injection as the email address
			_, err := sess.InsertInto("dbr_people").
				Pair("name", injectionAttempt).
				Exec()
			assert.NoError(t, err)

			// SELECT the name back and ensure it's equal to the injection attempt
			var name string
			err = sess.Select("name").From("dbr_people").OrderDir("id", false).Limit(1).LoadValue(&name)
			assert.Equal(t, injectionAttempt, name)
		}
	}
}

// InjectionAttempts is a newline separated list of common SQL injection exploits
// taken from https://wfuzz.googlecode.com/svn/trunk/wordlist/Injections/SQL.txt

const injectionAttempts = `
'
"
#
-
--
'%20--
--';
'%20;
=%20'
=%20;
=%20--
\x23
\x27
\x3D%20\x3B'
\x3D%20\x27
\x27\x4F\x52 SELECT *
\x27\x6F\x72 SELECT *
'or%20select *
admin'--
<>"'%;)(&+
'%20or%20''='
'%20or%20'x'='x
"%20or%20"x"="x
')%20or%20('x'='x
0 or 1=1
' or 0=0 --
" or 0=0 --
or 0=0 --
' or 0=0 #
" or 0=0 #
or 0=0 #
' or 1=1--
" or 1=1--
' or '1'='1'--
"' or 1 --'"
or 1=1--
or%201=1
or%201=1 --
' or 1=1 or ''='
" or 1=1 or ""="
' or a=a--
" or "a"="a
') or ('a'='a
") or ("a"="a
hi" or "a"="a
hi" or 1=1 --
hi' or 1=1 --
hi' or 'a'='a
hi') or ('a'='a
hi") or ("a"="a
'hi' or 'x'='x';
@variable
,@variable
PRINT
PRINT @@variable
select
insert
as
or
procedure
limit
order by
asc
desc
delete
update
distinct
having
truncate
replace
like
handler
bfilename
' or username like '%
' or uname like '%
' or userid like '%
' or uid like '%
' or user like '%
exec xp
exec sp
'; exec master..xp_cmdshell
'; exec xp_regread
t'exec master..xp_cmdshell 'nslookup www.google.com'--
--sp_password
\x27UNION SELECT
' UNION SELECT
' UNION ALL SELECT
' or (EXISTS)
' (select top 1
'||UTL_HTTP.REQUEST
1;SELECT%20*
to_timestamp_tz
tz_offset
&lt;&gt;&quot;'%;)(&amp;+
'%20or%201=1
%27%20or%201=1
%20$(sleep%2050)
%20'sleep%2050'
char%4039%41%2b%40SELECT
&apos;%20OR
'sqlattempt1
(sqlattempt2)
|
%7C
*|
%2A%7C
*(|(mail=*))
%2A%28%7C%28mail%3D%2A%29%29
*(|(objectclass=*))
%2A%28%7C%28objectclass%3D%2A%29%29
(
%28
)
%29
&
%26
!
%21
' or 1=1 or ''='
' or ''='
x' or 1=1 or 'x'='y
/
//
//*
*/*
`
