package fwschemadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

/*
The general algorithm would be calling the upstream tftypes.Transform()
function to walk each potential update value. Check if the value implements the
new interface, continuing to the next value without modification if not
implemented. If the update value is null or unknown, continue to the next value
without modification, meaning only known update values are now being considered.
Fetch the prior value matching the path of the update value. If the prior value
is null or unknown, continue to the next value without modification. Call the
interface method. If the response indicates to preserve the prior value, change
the value to the prior value and continue to the next value.
*/

// DoTheDew does ...
func (d *Data) DoTheDew(ctx context.Context, priorData Data) diag.Diagnostics {
	var diags diag.Diagnostics

	// Panic prevention
	if d == nil {
		return diags
	}

	// NOTE: A lot of the type system transformations could be avoided if the
	// framework type system re-implemented Walk() and Transform().
	updatedTerraformValue, err := tftypes.Transform(d.TerraformValue, func(tfTypePath *tftypes.AttributePath, tfTypeValue tftypes.Value) (tftypes.Value, error) {
		// Do not transform if the current value is null or unknown.
		if tfTypeValue.IsNull() || !tfTypeValue.IsKnown() {
			return tfTypeValue, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfTypePath, d.Schema)

		diags.Append(fwPathDiags...)

		// Do not transform if path cannot be converted.
		// Checking against fwPathDiags will capture all errors.
		if fwPathDiags.HasError() {
			return tfTypeValue, nil
		}

		priorAttrValue, priorAttrValueDiags := priorData.ValueAtPath(ctx, fwPath)

		diags.Append(priorAttrValueDiags...)

		// Do not transform if prior data cannot be fetched.
		// Checking against priorAttrValueDiags will capture all errors.
		if priorAttrValueDiags.HasError() {
			return tfTypeValue, nil
		}

		if priorAttrValue == nil {
			diags.AddAttributeError(
				fwPath,
				d.Description.Title()+" Read Error",
				"An unexpected error was encountered trying to read an attribute from the "+d.Description.String()+". This is always an error in the provider. Please report the following to the provider developer:\n\n"+
					"Missing attribute value, however no error was returned. Preventing the panic from this situation.\n\n"+
					"Path: "+fwPath.String(),
			)

			return tfTypeValue, nil
		}

		// Do not transform if prior data is null or unknown.
		if priorAttrValue.IsNull() || priorAttrValue.IsUnknown() {
			return tfTypeValue, nil
		}

		attrType, attrTypeDiags := d.Schema.TypeAtPath(ctx, fwPath)

		diags.Append(attrTypeDiags...)

		// Do not transform if the type cannot be determined.
		// Checking against attrTypeDiags will capture all errors.
		if attrTypeDiags.HasError() {
			return tfTypeValue, nil
		}

		currentAttrValue, err := attrType.ValueFromTerraform(ctx, tfTypeValue)

		if err != nil {
			diags.AddError(
				"Data Transformation Error",
				"An unexpected error occurred while transformating a current value from "+
					"the terraform-plugin-go type system to the terraform-plugin-framework type system. "+
					"Please report this to the provider developers.\n\n"+
					"Error: "+err.Error()+"\n"+
					"Path: "+fwPath.String(),
			)

			// nolint:nilerr // Use higher fidelity diagnostic above instead.
			return tfTypeValue, nil
		}

		var usePriorValue bool
		var semanticEqualsDiags diag.Diagnostics

		switch currentValue := currentAttrValue.(type) {
		case basetypes.StringValuableWithSemanticEquals:
			priorValue, ok := priorAttrValue.(basetypes.StringValuable)

			// Do not transform if the prior value does not match the expected
			// current type. Alternatively, this could raise an error to suggest
			// either a bug in this logic or to suggest the provider implement
			// UpgradeResourceState, but it would be hard to know the cause.
			if !ok {
				return tfTypeValue, nil
			}

			usePriorValue, semanticEqualsDiags = currentValue.StringSemanticEquals(ctx, priorValue)
		}

		diags.Append(semanticEqualsDiags...)

		// Do not transform if the semantic equality raised an error.
		// Checking against semanticEqualsDiags will capture all errors.
		if semanticEqualsDiags.HasError() {
			return tfTypeValue, nil
		}

		// Do not transform if the value semantic equality did not say to do so.
		if !usePriorValue {
			return tfTypeValue, nil
		}

		priorTfValue, err := priorAttrValue.ToTerraformValue(ctx)

		if err != nil {
			diags.AddError(
				"Data Transformation Error",
				"An unexpected error occurred while transformating a prior value from "+
					"the terraform-plugin-framework type system to the terraform-plugin-go type system. "+
					"Please report this to the provider developers.\n\n"+
					"Error: "+err.Error()+"\n"+
					"Path: "+fwPath.String(),
			)

			// nolint:nilerr // Use higher fidelity diagnostic above instead.
			return tfTypeValue, nil
		}

		return priorTfValue, nil
	})

	if err != nil {
		diags.AddError(
			"Data Transformation Error",
			"An unexpected error occurred while transforming data for type-based semantic equality. "+
				"This is always an error with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return diags
	}

	d.TerraformValue = updatedTerraformValue

	return diags
}
