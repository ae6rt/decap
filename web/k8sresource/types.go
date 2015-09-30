package k8sresource

import "speter.net/go/exp/math/dec/inf"

type Quantity struct {
	// Amount is public, so you can manipulate it if the accessor
	// functions are not sufficient.
	Amount *inf.Dec

	// Change Format at will. See the comment for Canonicalize for
	// more details.
	Format
}

type Format string
