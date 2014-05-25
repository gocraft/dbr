package dbr

import (
  "bytes"
  "fmt"
  "database/sql"
  "time"
)

type DeleteBuilder struct {
  *Session
  runner

  From string
  WhereFragments  []*whereFragment
  OrderBys        []string
  LimitCount      uint64
  LimitValid      bool
  OffsetCount     uint64
  OffsetValid     bool
}

func (sess *Session) DeleteFrom(from string) *DeleteBuilder {
  return &DeleteBuilder{
    Session: sess,
    runner:  sess.cxn.Db,
    From:    from,
  }
}

func (tx *Tx) DeleteFrom(from string) *DeleteBuilder {
  return &DeleteBuilder{
    Session: tx.Session,
    runner:  tx.Tx,
    From:    from,
  }
}

func (b *DeleteBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *DeleteBuilder {
  b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
  return b
}

func (b *DeleteBuilder) OrderBy(ord string) *DeleteBuilder {
  b.OrderBys = append(b.OrderBys, ord)
  return b
}

func (b *DeleteBuilder) OrderDir(ord string, isAsc bool) *DeleteBuilder {
  if isAsc {
    b.OrderBys = append(b.OrderBys, ord+" ASC")
  } else {
    b.OrderBys = append(b.OrderBys, ord+" DESC")
  }
  return b
}

func (b *DeleteBuilder) Limit(limit uint64) *DeleteBuilder {
  b.LimitCount = limit
  b.LimitValid = true
  return b
}

func (b *DeleteBuilder) Offset(offset uint64) *DeleteBuilder {
  b.OffsetCount = offset
  b.OffsetValid = true
  return b
}

func (b *DeleteBuilder) ToSql() (string, []interface{}) {
  if len(b.From) == 0 {
    panic("no table specified")
  }

  var sql bytes.Buffer
  var args []interface{}

  sql.WriteString("DELETE FROM ")
  sql.WriteString(b.From)

  // Write WHERE clause if we have any fragments
  if len(b.WhereFragments) > 0 {
    sql.WriteString(" WHERE ")
    writeWhereFragmentsToSql(b.WhereFragments, &sql, &args)
  }

  // Ordering and limiting
  if len(b.OrderBys) > 0 {
    sql.WriteString(" ORDER BY ")
    for i, s := range b.OrderBys {
      if i > 0 {
        sql.WriteString(", ")
      }
      sql.WriteString(s)
    }
  }

  if b.LimitValid {
    sql.WriteString(" LIMIT ")
    fmt.Fprint(&sql, b.LimitCount)
  }

  if b.OffsetValid {
    sql.WriteString(" OFFSET ")
    fmt.Fprint(&sql, b.OffsetCount)
  }

  return sql.String(), args
}

func (b *DeleteBuilder) Exec() (sql.Result, error) {
  sql, args := b.ToSql()

  fullSql, err := Interpolate(sql, args)
  if err != nil {
    return nil, b.EventErrKv("dbr.delete.exec.interpolate", err, kvs{"sql": fullSql})
  }

  // Start the timer:
  startTime := time.Now()
  defer func() { b.TimingKv("dbr.delete", time.Since(startTime).Nanoseconds(), kvs{"sql": fullSql}) }()

  result, err := b.runner.Exec(fullSql)
  if err != nil {
    return result, b.EventErrKv("dbr.delete.exec.exec", err, kvs{"sql": fullSql})
  }

  return result, nil
}
