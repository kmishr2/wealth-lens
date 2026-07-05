package finance

// FixedDepositValuationDefinition explains how fixed deposits participate in
// portfolio valuation without assuming an interest compounding convention.
func FixedDepositValuationDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Fixed deposit current value",
		Formula: "current value = ledger-derived fixed-deposit quantity × latest explicit fixed-deposit price",
		Assumptions: []string{
			"The fixed deposit is represented by one ledger asset unit at opening.",
			"Current value is supplied explicitly by the user or institution.",
			"Annual interest rate is contract metadata and is not used to estimate value.",
			"No compounding frequency, payout mode, penalty, tax, or currency conversion is assumed.",
		},
		RequiredInputs: []string{
			"ledger-derived fixed-deposit quantity",
			"latest explicit fixed-deposit price and valuation timestamp",
		},
		Explanation: "The displayed value is an auditable observation, not an accrued-interest estimate. Add a new immutable price observation when the institution reports a new value.",
	}
}
