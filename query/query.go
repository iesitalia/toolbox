package query

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/generic"
	"regexp"
	"slices"
	"strings"
)

// sortRegex is a regular expression used to parse and extract sorting information from a string.
// The regex pattern is "(?i)(?P<c1>[a-z_-]+)(\.(?P<c2>[a-z_-]+))?\.(?P<sort>asc|desc)".
// The pattern matches the following components:
// - "(?i)" - Case-insensitive matching.
// - "(?P<c1>[a-z_-]+)" - Captures a variable named "c1" that consists of one or more lowercase letters, underscores, or hyphens.
// - "(\.(?P<c2>[a-z_-]+))?)" - Optionally captures a variable named "c2" that consists of one or more lowercase letters, underscores, or hyphens, preceded by a dot.
// - "\.(?P<sort>asc|desc)" - Captures the sorting direction, either "asc" or "desc", preceded by a dot.
var sortRegex = regexp.MustCompile(`(?i)(?P<c1>[a-z_-]+)(\.(?P<c2>[a-z_-]+))?\.(?P<sort>asc|desc)`)

// Query represents a query for the database.
type Query struct {
	_select  []string
	_from    []string
	_where   []string
	_groupBy string
	_limit   string
	_offset  string
	_order   []string
	_joins   []*schema.Model
	raw      string
}

// Raw sets the raw query for the Query object.
// The raw query will be used instead of the generated query.
// Parameters:
//   - query: The raw query string to be set.
//
// Example usage:
//
//	q.Raw("SELECT * FROM users WHERE age > 25")
//
// Note: Setting a raw query will override any other query options that have been set,
// such as Select, From, Where, GroupBy, Order, Offset, and Limit.
func (q *Query) Raw(query string) {
	q.raw = query
}

// Select selects a column from the query's table to be included in the result set.
// An optional alias can be provided for the column.
func (q *Query) Select(column string, as ...string) error {
	var s = ""
	if len(as) > 0 {
		s = q.quoteSelect(column) + " AS `" + as[0] + "`"
	} else {
		s = q.quoteSelect(column)
	}

	if !slices.Contains(q._select, s) {
		q._select = append(q._select, s)
	}
	return nil
}

// quote trims the leading and trailing quotes from a string and
// wraps it with backticks. If the string contains a dot, it splits it
// into two chunks and wraps each chunk with backticks.
//
// If the string only contains one chunk, it wraps the entire string
// with backticks.
//
// Example:
//
//	s := quote("`test`") // returns "`test`"
//	s := quote("`test.string`") // returns "`test`.`string`"
func quote(s string) string {
	s = strings.Trim(s, "`'\"")
	var chunks = strings.Split(s, ".")
	if len(chunks) == 2 {
		return "`" + chunks[0] + "`." + "`" + chunks[1] + "`"
	}
	return "`" + s + "`"
}

// From appends a table name to the query's from clause.
// It quotes the table name and checks if it's already in the slice of from tables.
// If the table is not already in the slice, it adds it and calls schema.Find to find the corresponding model.
// If a model is found, it appends it to the slice of joins.
//
// Example usage:
//
//	func (q *Query) quoteSelect(s string) string {
//	  var chunks = strings.Split(s, ".")
//	  if len(chunks) == 2 {
//	      var table = strings.Trim(chunks[0], "`'\"")
//	      var column = strings.Trim(chunks[1], "`'\"")
//	      q.From(table)
//	      return "`" + table + "`." + "`" + column + "`"
//	  } else {
//	      s = strings.Trim(s, "`'\"")
//	  }
//	  return "`" + s + "`"
//	}
func (q *Query) From(s string) error {
	s = quote(s)
	if !slices.Contains(q._from, s) {
		if scm := schema.Find(strings.Trim(s, "`'\"")); scm != nil {
			q._joins = append(q._joins, scm)
		}
		q._from = append(q._from, s)
	}
	return nil
}

// Where adds a condition to the query's WHERE clause.
// The condition should be provided as a string.
// Multiple conditions can be added by calling this method multiple times.
func (q *Query) Where(s string) {
	q._where = append(q._where, s)
}

// GroupBy sets the GROUP BY clause in the query to the specified column or expression. This method is used to group the result set by one or more columns.
func (q *Query) GroupBy(s string) {
	q._groupBy = s
}

// Order splits the input string into chunks separated by commas.
// For each chunk, it extracts and matches patterns against the sortRegex.
// If there is a match and the length of the matches is 5,
// it constructs a new string by concatenating the matched pattern with space,
// and appends it to the _order slice of the Query struct.
// The method processes the input string to add ordering criteria to the query.
func (q *Query) Order(s string) {
	var chunks = strings.Split(s, ",")
	for _, item := range chunks {
		var matches = sortRegex.FindStringSubmatch(strings.TrimSpace(item))
		if len(matches) == 5 {
			if matches[2] != "" {
				s = quote(matches[1]+"."+matches[3]) + " " + matches[4]
			} else {
				s = quote(matches[1]) + " " + matches[4]
			}
			q._order = append(q._order, s)
		}
	}

}

// Offset sets the offset for the query to skip a specified number of rows before starting to return the rows.
// It takes a string as input representing the number of rows to skip.
// Example usage: query.Offset("10")
func (q *Query) Offset(s string) {
	q._offset = s
}

// Limit sets the number of rows to be returned in the query result.
//
// The `s` parameter specifies the limit as a string. The actual value of the limit is determined by parsing the string
// using the `generic.Parse` function and converting it to an int64 value.
//
// Example:
//
//	query.Limit("10") sets the limit to 10 rows.
//
// Note that the `Limit` method modifies the `_limit` field of the `Query` struct.
func (q *Query) Limit(s string) {
	q._limit = s
}

// GetCountQuery returns the SQL query string that retrieves the count of records matching the conditions specified in the Query object.
func (q *Query) GetCountQuery() string {
	var query = "SELECT COUNT(*) AS `count` FROM " + strings.Join(q._from, ",")
	var _, conditions, _ = q._joins[0].Join(q._joins[1:]...)
	q._where = append(q._where, conditions...)
	if len(q._where) > 0 {
		var condition = strings.TrimSpace(strings.Join(q._where, " AND "))
		if condition != "" {
			query += " WHERE " + condition
		}
	}
	if q._groupBy != "" {
		query += " GROUP BY " + q._groupBy
	}
	return query
}

// GetQuery returns the generated SQL query based on the current state of the Query object.
// If the raw query is set, it will be returned as is, without additional processing.
// Otherwise, the query will be constructed based on the selected columns, tables, where conditions, ordering,
// grouping, limit, and offset specified in the Query object.
func (q *Query) GetQuery() string {
	if q.raw != "" {
		return q.raw
	}
	var query = "SELECT " + strings.Join(q._select, ",") + " FROM " + strings.Join(q._from, ",")
	var _, conditions, _ = q._joins[0].Join(q._joins[1:]...)
	q._where = append(q._where, conditions...)
	if len(q._where) > 0 {
		var condition = strings.TrimSpace(strings.Join(q._where, " AND "))
		if condition != "" {
			query += " WHERE " + condition
		}
	}
	if q._groupBy != "" {
		query += " GROUP BY " + q._groupBy
	}

	if len(q._order) > 0 {
		query += " ORDER BY " + strings.Join(q._order, ",")
	}
	if q._limit != "" {
		query += " LIMIT " + fmt.Sprint(generic.Parse(q._limit).Int64())
	}
	if q._offset != "" {
		query += " OFFSET " + fmt.Sprint(generic.Parse(q._offset).Int64())
	}

	return query
}

// quoteSelect is a method of the Query struct that quotes a SELECT statement.
// It splits the given string by "." to extract the table and column names.
// If there are two chunks, it trims the backticks, single quotes, or double quotes from the table and column names,
// sets the table in the q.From method, and returns the quoted table.column string.
// Otherwise, it trims the backticks, single quotes, or double quotes from the string and returns the quoted string.
//
// Parameters:
// - s: The SELECT statement string.
//
// Returns:
// - The quoted SELECT statement string.
func (q *Query) quoteSelect(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 0 && s[0] == '(' && s[len(s)-1] == ')' {
		return s
	}
	var chunks = strings.Split(s, ".")
	if len(chunks) == 2 {
		var table = strings.Trim(chunks[0], "`'\"")
		var column = strings.Trim(chunks[1], "`'\"")
		q.From(table)
		return "`" + table + "`." + "`" + column + "`"
	} else {
		s = strings.Trim(s, "`'\"")
	}
	return "`" + s + "`"
}
