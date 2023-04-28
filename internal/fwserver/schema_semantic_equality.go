package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// SchemaSemanticEqualityRequest represents a request for a schema to run all
// semantic equality functions.
type SchemaSemanticEqualityRequest struct {
	// PriorData is the prior schema-based data.
	PriorData fwschemadata.Data

	// ProposedNewData is the proposed new schema-based data. The response
	// NewData contains the results of any modifications.
	ProposedNewData fwschemadata.Data
}

// SchemaSemanticEqualityResponse represents a response to a SchemaSemanticEqualityRequest.
type SchemaSemanticEqualityResponse struct {
	// NewData is the new schema-based data after any modifications.
	NewData fwschemadata.Data

	// Diagnostics report errors or warnings related to running all attribute
	// plan modifiers. Returning an empty slice indicates a successful
	// plan modification with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// SchemaSemanticEquality runs all semantic equality logic in all schema
// attributes and blocks.
func SchemaSemanticEquality(ctx context.Context, req SchemaSemanticEqualityRequest, resp *SchemaSemanticEqualityResponse) {
	var diags diag.Diagnostics

	for name, attribute := range req.ProposedNewData.Schema.GetAttributes() {
		attrReq := AttributeSemanticEqualityRequest{
			Attribute:                   attribute,
			Path:                        path.Root(name),
			PriorValueDescription:       req.PriorData.Description,
			ProposedNewValueDescription: req.ProposedNewData.Description,
		}

		attrReq.PriorValue, diags = req.PriorData.ValueAtPath(ctx, attrReq.Path)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		attrReq.ProposedNewValue, diags = req.ProposedNewData.ValueAtPath(ctx, attrReq.Path)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		attrResp := AttributeSemanticEqualityResponse{
			NewValue: attrReq.ProposedNewValue,
		}

		AttributeSemanticEquality(ctx, attrReq, &attrResp)

		resp.Diagnostics.Append(attrResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}

		if attrResp.NewValue.Equal(attrReq.ProposedNewValue) {
			continue
		}

		resp.Diagnostics.Append(resp.NewData.SetAtPath(ctx, attrReq.Path, attrResp.NewValue)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	for name, block := range req.ProposedNewData.Schema.GetBlocks() {
		blockReq := BlockSemanticEqualityRequest{
			Block:                       block,
			Path:                        path.Root(name),
			PriorValueDescription:       req.PriorData.Description,
			ProposedNewValueDescription: req.ProposedNewData.Description,
		}

		blockReq.PriorValue, diags = req.PriorData.ValueAtPath(ctx, blockReq.Path)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		blockReq.ProposedNewValue, diags = req.ProposedNewData.ValueAtPath(ctx, blockReq.Path)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		blockResp := BlockSemanticEqualityResponse{
			NewValue: blockReq.ProposedNewValue,
		}

		BlockSemanticEquality(ctx, blockReq, &blockResp)

		resp.Diagnostics.Append(blockResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}

		if blockResp.NewValue.Equal(blockReq.ProposedNewValue) {
			continue
		}

		resp.Diagnostics.Append(resp.NewData.SetAtPath(ctx, blockReq.Path, blockResp.NewValue)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}
}
