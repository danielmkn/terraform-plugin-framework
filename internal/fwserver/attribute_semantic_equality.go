package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// AttributeSemanticEqualityRequest represents a request for the provider to
// perform semantic equality logic on an attribute value.
type AttributeSemanticEqualityRequest struct {
	// Attribute is the definition of the attribute.
	Attribute fwschema.Attribute

	// Path is the schema-based path of the attribute.
	Path path.Path

	// PriorValue is the prior value of the attribute.
	PriorValue attr.Value

	// PriorValueDescription is the data description of the prior value to
	// enhance diagnostics.
	PriorValueDescription fwschemadata.DataDescription

	// ProposedNewValue is the proposed new value of the attribute. NewValue in
	// the response contains the results of semantic equality logic.
	ProposedNewValue attr.Value

	// ProposedNewValueDescription is the data description of the proposed new
	// value to enhance diagnostics.
	ProposedNewValueDescription fwschemadata.DataDescription
}

type AttributeSemanticEqualityResponse struct {
	// NewValue contains the new value of the attribute based on the semantic
	// equality logic.
	NewValue attr.Value

	// Diagnostics contains any errors and warnings for the logic.
	Diagnostics diag.Diagnostics
}

// AttributeSemanticEquality runs all semantic equality logic for an attribute.
func AttributeSemanticEquality(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.Path.String())

	// Ensure the response NewValue always starts with the proposed new value.
	resp.NewValue = req.ProposedNewValue

	// If the prior value is null or unknown, no need to check semantic equality
	// as the proposed new value is always correct. There is also no need to
	// descend further into any nesting.
	if req.PriorValue.IsNull() || req.PriorValue.IsUnknown() {
		return
	}

	// If the proposed new value is null or unknown, no need to check semantic
	// equality as it should never be changed back to the prior value. There is
	// also no need to descend further into any nesting.
	if req.ProposedNewValue.IsNull() || req.ProposedNewValue.IsUnknown() {
		return
	}

	switch req.Attribute.GetType().(type) {
	case basetypes.BoolTypable:
		AttributeSemanticEqualityBool(ctx, req, resp)
	case basetypes.Float64Typable:
		AttributeSemanticEqualityFloat64(ctx, req, resp)
	case basetypes.Int64Typable:
		AttributeSemanticEqualityInt64(ctx, req, resp)
	case basetypes.ListTypable:
		AttributeSemanticEqualityList(ctx, req, resp)
	case basetypes.MapTypable:
		AttributeSemanticEqualityMap(ctx, req, resp)
	case basetypes.NumberTypable:
		AttributeSemanticEqualityNumber(ctx, req, resp)
	case basetypes.ObjectTypable:
		AttributeSemanticEqualityObject(ctx, req, resp)
	case basetypes.SetTypable:
		AttributeSemanticEqualitySet(ctx, req, resp)
	case basetypes.StringTypable:
		AttributeSemanticEqualityString(ctx, req, resp)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	nestedAttribute, ok := req.Attribute.(fwschema.NestedAttribute)

	if !ok {
		return
	}

	nestedAttributeObject := nestedAttribute.GetNestedObject()

	nm := nestedAttribute.GetNestingMode()
	switch nm {
	case fwschema.NestingModeList:
		priorList, diags := coerceListValue(ctx, req.Path, req.PriorValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		// Use response as the planned value may have been modified with list
		// semantic equality.
		proposedNewList, diags := coerceListValue(ctx, req.Path, resp.NewValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		newValueElements := proposedNewList.Elements()

		for idx, planElem := range newValueElements {
			attrPath := req.Path.AtListIndex(idx)

			priorObject, diags := listElemObject(ctx, attrPath, priorList, idx, req.PriorValueDescription)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			proposedNewObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			newObjectAttributes := proposedNewObject.Attributes()

			for nestedName, nestedAttr := range nestedAttributeObject.GetAttributes() {
				nestedPriorValue, diags := objectAttributeValue(ctx, priorObject, nestedName, req.PriorValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedProposedNewValue, diags := objectAttributeValue(ctx, proposedNewObject, nestedName, req.ProposedNewValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedAttrReq := AttributeSemanticEqualityRequest{
					Attribute:                   nestedAttr,
					Path:                        req.Path.AtName(nestedName),
					PriorValue:                  nestedPriorValue,
					PriorValueDescription:       req.PriorValueDescription,
					ProposedNewValue:            nestedProposedNewValue,
					ProposedNewValueDescription: req.ProposedNewValueDescription,
				}
				nestedAttrResp := &AttributeSemanticEqualityResponse{
					NewValue: nestedAttrReq.ProposedNewValue,
				}

				AttributeSemanticEquality(ctx, nestedAttrReq, nestedAttrResp)

				newObjectAttributes[nestedName] = nestedAttrResp.NewValue
				resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
			}

			newObject, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newObjectAttributes)

			resp.Diagnostics.Append(diags...)

			newValueElements[idx] = newObject
		}

		newValue, diags := types.ListValue(proposedNewList.ElementType(ctx), newValueElements)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	case fwschema.NestingModeSet:
		priorSet, diags := coerceSetValue(ctx, req.Path, req.PriorValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		// Use response as the planned value may have been modified with set
		// semantic equality.
		proposedNewSet, diags := coerceSetValue(ctx, req.Path, resp.NewValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		newValueElements := proposedNewSet.Elements()

		for idx, planElem := range newValueElements {
			attrPath := req.Path.AtSetValue(planElem)

			priorObject, diags := setElemObject(ctx, attrPath, priorSet, idx, req.PriorValueDescription)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			proposedNewObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			newObjectAttributes := proposedNewObject.Attributes()

			for nestedName, nestedAttr := range nestedAttributeObject.GetAttributes() {
				nestedPriorValue, diags := objectAttributeValue(ctx, priorObject, nestedName, req.PriorValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedProposedNewValue, diags := objectAttributeValue(ctx, proposedNewObject, nestedName, req.ProposedNewValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedAttrReq := AttributeSemanticEqualityRequest{
					Attribute:                   nestedAttr,
					Path:                        req.Path.AtName(nestedName),
					PriorValue:                  nestedPriorValue,
					PriorValueDescription:       req.PriorValueDescription,
					ProposedNewValue:            nestedProposedNewValue,
					ProposedNewValueDescription: req.ProposedNewValueDescription,
				}
				nestedAttrResp := &AttributeSemanticEqualityResponse{
					NewValue: nestedAttrReq.ProposedNewValue,
				}

				AttributeSemanticEquality(ctx, nestedAttrReq, nestedAttrResp)

				newObjectAttributes[nestedName] = nestedAttrResp.NewValue
				resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
			}

			newObject, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newObjectAttributes)

			resp.Diagnostics.Append(diags...)

			newValueElements[idx] = newObject
		}

		newValue, diags := types.SetValue(proposedNewSet.ElementType(ctx), newValueElements)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	case fwschema.NestingModeMap:
		priorMap, diags := coerceMapValue(ctx, req.Path, req.PriorValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		// Use response as the planned value may have been modified with map
		// semantic equality.
		proposedNewMap, diags := coerceMapValue(ctx, req.Path, resp.NewValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		newValueElements := proposedNewMap.Elements()

		for key, planElem := range newValueElements {
			attrPath := req.Path.AtMapKey(key)

			priorObject, diags := mapElemObject(ctx, attrPath, priorMap, key, req.PriorValueDescription)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			proposedNewObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			newObjectAttributes := proposedNewObject.Attributes()

			for nestedName, nestedAttr := range nestedAttributeObject.GetAttributes() {
				nestedPriorValue, diags := objectAttributeValue(ctx, priorObject, nestedName, req.PriorValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedProposedNewValue, diags := objectAttributeValue(ctx, proposedNewObject, nestedName, req.ProposedNewValueDescription)

				resp.Diagnostics.Append(diags...)

				if diags.HasError() {
					return
				}

				nestedAttrReq := AttributeSemanticEqualityRequest{
					Attribute:                   nestedAttr,
					Path:                        req.Path.AtName(nestedName),
					PriorValue:                  nestedPriorValue,
					PriorValueDescription:       req.PriorValueDescription,
					ProposedNewValue:            nestedProposedNewValue,
					ProposedNewValueDescription: req.ProposedNewValueDescription,
				}
				nestedAttrResp := &AttributeSemanticEqualityResponse{
					NewValue: nestedAttrReq.ProposedNewValue,
				}

				AttributeSemanticEquality(ctx, nestedAttrReq, nestedAttrResp)

				newObjectAttributes[nestedName] = nestedAttrResp.NewValue
				resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
			}

			newObject, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newObjectAttributes)

			resp.Diagnostics.Append(diags...)

			newValueElements[key] = newObject
		}

		newValue, diags := types.MapValue(proposedNewMap.ElementType(ctx), newValueElements)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	case fwschema.NestingModeSingle:
		priorObject, diags := coerceObjectValue(ctx, req.Path, req.PriorValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		// Use response as the planned value may have been modified with object
		// semantic equality.
		proposedNewObject, diags := coerceObjectValue(ctx, req.Path, resp.NewValue)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		newValueAttributes := proposedNewObject.Attributes()

		for nestedName, nestedAttr := range nestedAttributeObject.GetAttributes() {
			nestedPriorValue, diags := objectAttributeValue(ctx, priorObject, nestedName, req.PriorValueDescription)

			resp.Diagnostics.Append(diags...)

			if diags.HasError() {
				return
			}

			nestedProposedNewValue, diags := objectAttributeValue(ctx, proposedNewObject, nestedName, req.ProposedNewValueDescription)

			resp.Diagnostics.Append(diags...)

			if diags.HasError() {
				return
			}

			nestedAttrReq := AttributeSemanticEqualityRequest{
				Attribute:                   nestedAttr,
				Path:                        req.Path.AtName(nestedName),
				PriorValue:                  nestedPriorValue,
				PriorValueDescription:       req.PriorValueDescription,
				ProposedNewValue:            nestedProposedNewValue,
				ProposedNewValueDescription: req.ProposedNewValueDescription,
			}
			nestedAttrResp := &AttributeSemanticEqualityResponse{
				NewValue: nestedAttrReq.ProposedNewValue,
			}

			AttributeSemanticEquality(ctx, nestedAttrReq, nestedAttrResp)

			newValueAttributes[nestedName] = nestedAttrResp.NewValue
			resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
		}

		newValue, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newValueAttributes)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Attribute Semantic Equality Error",
			"An unexpected error occurred while walking the schema for attribute semantic equality. "+
				"This is always an error with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				"Error: Unknown attribute nesting mode "+fmt.Sprintf("(%T: %v)", nm, nm)+
				"Path: "+req.Path.String(),
		)

		return
	}
}

// AttributeSemanticEqualityBool performs all types.Bool semantic equality.
func AttributeSemanticEqualityBool(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.BoolValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.BoolValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.BoolSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityFloat64 performs all types.Float64 semantic equality.
func AttributeSemanticEqualityFloat64(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.Float64ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.Float64ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.Float64SemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityInt64 performs all types.Int64 semantic equality.
func AttributeSemanticEqualityInt64(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.Int64ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.Int64ValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.Int64SemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityList performs all types.List semantic equality.
func AttributeSemanticEqualityList(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	// TODO: Loop through all underlying elements
	priorValuable, ok := req.PriorValue.(basetypes.ListValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.ListValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.ListSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityMap performs all types.Map semantic equality.
func AttributeSemanticEqualityMap(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	// TODO: Loop through all underlying elements
	priorValuable, ok := req.PriorValue.(basetypes.MapValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.MapValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.MapSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityNumber performs all types.Number semantic equality.
func AttributeSemanticEqualityNumber(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.NumberValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.NumberValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.NumberSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityObject performs all types.Object semantic equality.
func AttributeSemanticEqualityObject(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	// TODO: Loop through all underlying attributes
	priorValuable, ok := req.PriorValue.(basetypes.ObjectValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.ObjectValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.ObjectSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualitySet performs all types.Set semantic equality.
func AttributeSemanticEqualitySet(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	// TODO: Loop through all underlying elements
	priorValuable, ok := req.PriorValue.(basetypes.SetValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.SetValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.SetSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}

// AttributeSemanticEqualityString performs all types.String semantic equality.
func AttributeSemanticEqualityString(ctx context.Context, req AttributeSemanticEqualityRequest, resp *AttributeSemanticEqualityResponse) {
	priorValuable, ok := req.PriorValue.(basetypes.StringValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	proposedNewValuable, ok := req.ProposedNewValue.(basetypes.StringValuableWithSemanticEquals)

	// No changes required if the interface is not implemented.
	if !ok {
		return
	}

	logging.FrameworkTrace(
		ctx,
		"Calling provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	usePriorValue, diags := proposedNewValuable.StringSemanticEquals(ctx, priorValuable)

	logging.FrameworkTrace(
		ctx,
		"Called provider defined type-based SemanticEquals",
		map[string]interface{}{
			logging.KeyValueType: proposedNewValuable.String(),
		},
	)

	resp.Diagnostics.Append(diags...)

	if !usePriorValue {
		return
	}

	resp.NewValue = priorValuable
}
