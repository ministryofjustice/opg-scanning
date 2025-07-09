package ingestion

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestSetValidatorValidationRules(t *testing.T) {
	tests := []struct {
		Name         string
		Set          types.BaseSet
		ErrorMessage string
	}{
		{
			Name: "missing <Header>",
			Set: types.BaseSet{
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1F", NoPages: 14},
					},
				},
			},
			ErrorMessage: "missing required Header element",
		},
		{
			Name: "missing schedule attribute",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1F", NoPages: 14},
					},
				},
			},
			ErrorMessage: "missing required Schedule attribute on Header",
		},
		{
			Name: "missing documents",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
			},
			ErrorMessage: "no Document elements found in Body",
		},
		{
			Name: "document missing type",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{NoPages: 14},
					},
				},
			},
			ErrorMessage: "document Type attribute is missing",
		},
		{
			Name: "document missing NoPages",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1H"},
					},
				},
			},
			ErrorMessage: "document NoPages attribute is missing or invalid",
		},
		{
			Name: "creating case with CaseNo set",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
					CaseNo:   "7000-0238-2394",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1H", NoPages: 14},
					},
				},
			},
			ErrorMessage: "must not supply a case number when creating a new case",
		},
		{
			Name: "adding correspondence without CaseNo set",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "Correspondence", NoPages: 5},
					},
				},
			},
			ErrorMessage: "must supply a case number when not creating a new case",
		},
		{
			Name: "multiple case creators in one set",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1H", NoPages: 14},
						{Type: "LP1F", NoPages: 16},
					},
				},
			},
			ErrorMessage: "set cannot contain multiple cases which would create a case",
		},
		{
			Name: "accepts new case document without CaseNo",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1H", NoPages: 14},
					},
				},
			},
			ErrorMessage: "",
		},
		{
			Name: "accepts correspondence with CaseNo",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
					CaseNo:   "7000-0238-2394",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "Correspondence", NoPages: 5},
					},
				},
			},
			ErrorMessage: "",
		},
		{
			Name: "accepts mixed sets",
			Set: types.BaseSet{
				Header: &types.BaseHeader{
					Schedule: "schedule-id",
				},
				Body: types.BaseBody{
					Documents: []types.BaseDocument{
						{Type: "LP1H", NoPages: 14},
						{Type: "Correspondence", NoPages: 5},
					},
				},
			},
			ErrorMessage: "",
		},
	}

	v := NewValidator()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			err := v.ValidateSet(&tc.Set)

			if tc.ErrorMessage != "" {
				assert.Equal(t, tc.ErrorMessage, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
