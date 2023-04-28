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

// BlockSemanticEqualityRequest represents a request for the provider to
// perform semantic equality logic on a block value.
type BlockSemanticEqualityRequest struct {
	// Block is the definition of the block.
	Block fwschema.Block

	// Path is the schema-based path of the block.
	Path path.Path

	// PriorValue is the prior value of the block.
	PriorValue attr.Value

	// PriorValueDescription is the data description of the prior value to
	// enhance diagnostics.
	PriorValueDescription fwschemadata.DataDescription

	// ProposedNewValue is the proposed new value of the block. NewValue in
	// the response contains the results of semantic equality logic.
	ProposedNewValue attr.Value

	// ProposedNewValueDescription is the data description of the proposed new
	// value to enhance diagnostics.
	ProposedNewValueDescription fwschemadata.DataDescription
}

type BlockSemanticEqualityResponse struct {
	// NewValue contains the new value of the block based on the semantic
	// equality logic.
	NewValue attr.Value

	// Diagnostics contains any errors and warnings for the logic.
	Diagnostics diag.Diagnostics
}

// BlockSemanticEquality runs all semantic equality logic for a block.
func BlockSemanticEquality(ctx context.Context, req BlockSemanticEqualityRequest, resp *BlockSemanticEqualityResponse) {
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

	// Using the existing Attribute logic to simplify the implementation. If
	// Attribute and Block ever support semantic equality logic at that level,
	// Block-based functions will need to be created.
	attrReq := AttributeSemanticEqualityRequest{
		Path:                        req.Path,
		PriorValue:                  req.PriorValue,
		PriorValueDescription:       req.PriorValueDescription,
		ProposedNewValue:            req.ProposedNewValue,
		ProposedNewValueDescription: req.ProposedNewValueDescription,
	}
	attrResp := &AttributeSemanticEqualityResponse{
		NewValue: attrReq.ProposedNewValue,
	}

	switch req.Block.Type().(type) {
	case basetypes.ListTypable:
		AttributeSemanticEqualityList(ctx, attrReq, attrResp)
	case basetypes.ObjectTypable:
		AttributeSemanticEqualityObject(ctx, attrReq, attrResp)
	case basetypes.SetTypable:
		AttributeSemanticEqualitySet(ctx, attrReq, attrResp)
	}

	resp.Diagnostics.Append(attrResp.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.NewValue = attrResp.NewValue

	nestedBlockObject := req.Block.GetNestedObject()

	nm := req.Block.GetNestingMode()
	switch nm {
	case fwschema.BlockNestingModeList:
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

			for nestedName, nestedAttr := range nestedBlockObject.GetAttributes() {
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

			for nestedName, nestedBlock := range nestedBlockObject.GetBlocks() {
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

				nestedBlockReq := BlockSemanticEqualityRequest{
					Block:                       nestedBlock,
					Path:                        req.Path.AtName(nestedName),
					PriorValue:                  nestedPriorValue,
					PriorValueDescription:       req.PriorValueDescription,
					ProposedNewValue:            nestedProposedNewValue,
					ProposedNewValueDescription: req.ProposedNewValueDescription,
				}
				nestedBlockResp := &BlockSemanticEqualityResponse{
					NewValue: nestedBlockReq.ProposedNewValue,
				}

				BlockSemanticEquality(ctx, nestedBlockReq, nestedBlockResp)

				newObjectAttributes[nestedName] = nestedBlockResp.NewValue
				resp.Diagnostics.Append(nestedBlockResp.Diagnostics...)
			}

			newObject, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newObjectAttributes)

			resp.Diagnostics.Append(diags...)

			newValueElements[idx] = newObject
		}

		newValue, diags := types.ListValue(proposedNewList.ElementType(ctx), newValueElements)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	case fwschema.BlockNestingModeSet:
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

			for nestedName, nestedAttr := range nestedBlockObject.GetAttributes() {
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

			for nestedName, nestedBlock := range nestedBlockObject.GetBlocks() {
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

				nestedBlockReq := BlockSemanticEqualityRequest{
					Block:                       nestedBlock,
					Path:                        req.Path.AtName(nestedName),
					PriorValue:                  nestedPriorValue,
					PriorValueDescription:       req.PriorValueDescription,
					ProposedNewValue:            nestedProposedNewValue,
					ProposedNewValueDescription: req.ProposedNewValueDescription,
				}
				nestedBlockResp := &BlockSemanticEqualityResponse{
					NewValue: nestedBlockReq.ProposedNewValue,
				}

				BlockSemanticEquality(ctx, nestedBlockReq, nestedBlockResp)

				newObjectAttributes[nestedName] = nestedBlockResp.NewValue
				resp.Diagnostics.Append(nestedBlockResp.Diagnostics...)
			}

			newObject, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newObjectAttributes)

			resp.Diagnostics.Append(diags...)

			newValueElements[idx] = newObject
		}

		newValue, diags := types.SetValue(proposedNewSet.ElementType(ctx), newValueElements)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	case fwschema.BlockNestingModeSingle:
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

		for nestedName, nestedAttr := range nestedBlockObject.GetAttributes() {
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

		for nestedName, nestedBlock := range nestedBlockObject.GetBlocks() {
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

			nestedBlockReq := BlockSemanticEqualityRequest{
				Block:                       nestedBlock,
				Path:                        req.Path.AtName(nestedName),
				PriorValue:                  nestedPriorValue,
				PriorValueDescription:       req.PriorValueDescription,
				ProposedNewValue:            nestedProposedNewValue,
				ProposedNewValueDescription: req.ProposedNewValueDescription,
			}
			nestedBlockResp := &BlockSemanticEqualityResponse{
				NewValue: nestedBlockReq.ProposedNewValue,
			}

			BlockSemanticEquality(ctx, nestedBlockReq, nestedBlockResp)

			newValueAttributes[nestedName] = nestedBlockResp.NewValue
			resp.Diagnostics.Append(nestedBlockResp.Diagnostics...)
		}

		newValue, diags := types.ObjectValue(proposedNewObject.AttributeTypes(ctx), newValueAttributes)

		resp.Diagnostics.Append(diags...)

		resp.NewValue = newValue
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Block Semantic Equality Error",
			"An unexpected error occurred while walking the schema for block semantic equality. "+
				"This is always an error with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				"Error: Unknown block nesting mode "+fmt.Sprintf("(%T: %v)", nm, nm)+
				"Path: "+req.Path.String(),
		)

		return
	}
}
