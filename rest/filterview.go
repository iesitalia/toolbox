package rest

import (
	"fmt"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db"
	scm "github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/tpl"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"toolbox"
	"toolbox/query"
)

// Join represents a join operation in a database query.
type Join struct {
	Table     string
	Condition string
}

// Select represents a selection in a database query.
type Select struct {
	Select string
	As     string
}

// FilterView represents a data structure for defining filter views.
type FilterView struct {
	Model       schema.Tabler      `json:"-"`
	Title       string             `json:"title,omitempty" json:"title,omitempty"`
	Description string             `json:"description,omitempty" json:"description,omitempty"`
	Select      []Select           `json:"-" json:"select,omitempty" json:"select,omitempty"`
	Join        []Join             `json:"-" json:"join,omitempty"`
	Condition   []string           `json:"condition,omitempty" json:"condition,omitempty"`
	Order       []string           `json:"order,omitempty" json:"order,omitempty"`
	URLParams   []Filter           `json:"url_params"`
	Filters     []Filter           `json:"filters,omitempty" json:"filters,omitempty"`
	Columns     []FilterViewColumn `json:"columns,omitempty" json:"columns,omitempty"`
}

// FilterViewColumn represents a column in a filter view. It has properties such as title, href, type, processor, sort, options, dbField, and actions.
// - Title: The title of the column.
// - Href: The href of the column.
// - Type: The type of the column.
// - Processor: A function that processes the data of the column.
// - Sort: A flag indicating whether the column can be sorted.
// - Options: The list of options for the column.
// - DBField: The database field of the column.
// - Actions: The list of actions for the column.
type FilterViewColumn struct {
	Title     string                                   `json:"title,omitempty"`
	Href      string                                   `json:"href,omitempty"`
	Type      string                                   `json:"type"`
	Processor func(data map[string]interface{}) string `json:"-"`
	Sort      bool                                     `json:"sort"`
	Options   toolbox.Dictionary[string]               `json:"list,omitempty"`
	DBField   string                                   `json:"-"`
	Actions   []Action                                 `json:"-"`
}

// Filter represents a filter for data retrieval.
// It contains the following properties:
//
// - Title: the title of the filter.
// - Type: the type of the filter.
// - Options: dictionary of options for the filter.
// - Name: the name of the filter.
// - Filter: the filter condition to be applied.
type Filter struct {
	Title   string                     `json:"title,omitempty"`
	Type    string                     `json:"type,omitempty"`
	Options toolbox.Dictionary[string] `json:"options,omitempty"`
	Name    string                     `json:"name,omitempty"`
	Filter  string                     `json:"-"`
}

// Action represents an action that can be performed in a view.
type Action struct {
	Type    string `json:"type,omitempty"`
	Href    string `json:"href,omitempty"`
	OnClick string `json:"on_click,omitempty"`
	Text    string `json:"text,omitempty"`
	Icon    string `json:"icon,omitempty"`
}

// GetData retrieves data from the FilterView based on the given offset, size, and request parameters. It returns an error if the model or the queries are invalid, the total number of
func (v *FilterView) GetData(offset int, size int, request *evo.Request) (error, int64, [][]interface{}) {
	var query = query.Query{}
	for _, item := range v.Columns {
		if item.DBField == "-" || item.DBField == "" {
			continue
		}
		query.Select(item.DBField)
	}

	var order = request.Query("sort").String()
	if order != "" {
		valid := true
		for _, item := range strings.Split(order, ",") {
			if !orderRegex.MatchString(item) {
				valid = false
				break
			}
		}
		if valid {
			query.Order(order)
		}
	}

	if v.Model == nil {
		return fmt.Errorf("invalid model %s", reflect.TypeOf(v.Model).Name()), 0, nil
	}
	m := scm.Find(v.Model.TableName())
	if m == nil {
		return fmt.Errorf("invalid model %s", reflect.TypeOf(v.Model).Name()), 0, nil
	}
	query.Select(m.Table+"."+m.PrimaryKey[0], "pk")
	for _, item := range v.Select {
		if item.As != "" {
			query.Select(item.Select, item.As)
		} else {
			query.Select(item.Select)
		}
	}

	query.From(v.Model.TableName())
	for _, item := range v.Join {
		query.From(item.Table)
		if item.Condition != "" {
			query.Where(item.Condition)
		}
	}

	for _, item := range v.URLParams {
		query.Where(strings.Replace(item.Filter, "*", request.Param(item.Name).String(), -1))
	}

	query.Limit(fmt.Sprint(size))
	query.Offset(fmt.Sprint(offset))
	for _, item := range v.Filters {
		if request.Query(item.Name).String() != "" {
			query.Where(strings.Replace(item.Filter, "*", request.Query(item.Name).String(), -1))
		}
	}

	var total int64
	db.Raw(query.GetCountQuery()).Scan(&total)

	var data []map[string]interface{}
	db.Debug().Raw(query.GetQuery()).Scan(&data)

	var result = make([][]interface{}, len(data))

	for j, row := range data {
		var item = make([]interface{}, len(v.Columns))
		for i, column := range v.Columns {
			if len(column.Actions) > 0 {
				var buttons []Action
				for _, action := range column.Actions {
					buttons = append(buttons, Action{
						Type:    action.Type,
						Href:    tpl.Render(action.Href, row),
						OnClick: tpl.Render(action.OnClick, row),
						Text:    action.Text,
						Icon:    action.Icon,
					})
				}
				item[i] = buttons
				continue
			}
			if column.Processor == nil {
				item[i] = fmt.Sprint(row[column.DBField])

			} else {
				item[i] = column.Processor(row)
			}
			if column.Href != "" {
				item[i] = "<a href=\"" + tpl.Render(column.Href, row) + "\">" + fmt.Sprint(item[i]) + "</a>"
			}

		}
		result[j] = item
	}

	return nil, total, result
}

// SetSelect adds the given Select to the FilterView's Select field
func (v *FilterView) SetSelect(s Select) {
	var skip = false
	for _, item := range v.Select {
		if item.Select == s.Select {
			skip = true
			break
		}
	}
	if !skip {
		v.Select = append(v.Select, s)
	}
}
