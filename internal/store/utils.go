package store

import "fmt"

func addFilter(column, operator string, value any, conditions *[]string, args *[]any) {
	*args = append(*args, value)
	*conditions = append(*conditions, fmt.Sprintf("%s %s $%d", column, operator, len(*args)))
}
