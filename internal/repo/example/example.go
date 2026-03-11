package growth

import (
	"context"
	"mytemplate/internal/database"
	model "mytemplate/internal/models/example"
)

// CreateExample insert example data into database
func CreateExample(ctx *context.Context, activity *model.ExampleTable) error {
	if activity == nil {
		return nil
	}
	return database.GetDBMysqlExample().Create(activity).Error
}
