package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDoTheDew(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          *fwschemadata.Data
		priorData     fwschemadata.Data
		expected      *fwschemadata.Data
		expectedDiags diag.Diagnostics
	}{
		"nil-data": {
			data: nil,
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: nil,
		},
		"misaligned-prior-data-attribute": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"not_test": testschema.Attribute{ // intentionally different name
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"not_test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"not_test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"State Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. "+
						"This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"test\") still remains in the path: could not find attribute or block \"test\" in schema",
				),
			},
		},
		"misaligned-prior-data-attribute-type": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.BoolType, // intentionally not StringType
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			expectedDiags: nil, // this should likely error, but for some reason it is not
		},
		"null-data": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					nil,
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					nil,
				),
			},
		},
		"null-prior-data": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					nil,
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
		},
		"null-value-current": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
		},
		"null-value-prior": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
		},
		"unknown-value-current": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
		},
		"unknown-value-prior": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
		},
		"list-nested-attribute-without-semantic-equality": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-with-semantic-equality-true": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
		},
		"list-nested-attribute-with-semantic-equality-false": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_nested_attribute": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_nested_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-without-semantic-equality": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "current"),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "prior"),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "current"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-with-semantic-equality-true": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "current"),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "prior"),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "prior"),
							},
						),
					},
				),
			},
		},
		"single-nested-attribute-with-semantic-equality-false": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "current"),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "prior"),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"single_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_nested_attribute": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_nested_attribute": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"test": tftypes.NewValue(tftypes.String, "current"),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-without-semantic-equality": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-with-semantic-equality-true": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: true,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										// TODO: Move logic and testing to fwserver for
										// set element alignment.
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
		},
		"set-nested-attribute-with-semantic-equality-false": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "prior"),
									},
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_nested_attribute": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"test": testschema.Attribute{
										Type: testtypes.StringTypeWithSemanticEquals{
											SemanticEquals: false,
										},
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_nested_attribute": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_nested_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test": tftypes.NewValue(tftypes.String, "current"),
									},
								),
							},
						),
					},
				),
			},
		},
		"string-without-semantic-equality": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
		},
		"string-with-semantic-equality-true": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
		},
		"string-with-semantic-equality-false": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
			priorData: fwschemadata.Data{
				Description: fwschemadata.DataDescriptionPlan,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "prior"),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionState,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "current"),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.DoTheDew(context.Background(), testCase.priorData)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("%s: %s:\n%s\n", d.Severity(), d.Summary(), d.Detail())
				}

				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
