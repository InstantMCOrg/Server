package models

import (
	"github.com/instantminecraft/server/pkg/enums"
	"testing"
)

func TestMcContainerSearchConfigDefaultStatus(t *testing.T) {
	defaultValues := McContainerSearchConfig{}
	if defaultValues.Status != enums.Prepared {
		t.Errorf("Default Status is %s but it should be %s", defaultValues.Status, enums.Prepared)
	}
}
