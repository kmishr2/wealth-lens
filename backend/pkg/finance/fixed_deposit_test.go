package finance

import "testing"

func TestFixedDepositValuationDefinitionDisclosesNoAccrualAssumption(t *testing.T) {
	definition := FixedDepositValuationDefinition()
	if definition.Name == "" || definition.Formula == "" || definition.Explanation == "" {
		t.Fatal("fixed deposit valuation definition is incomplete")
	}
	if len(definition.RequiredInputs) != 2 || len(definition.Assumptions) < 4 {
		t.Fatalf("definition inputs/assumptions = %d/%d", len(definition.RequiredInputs), len(definition.Assumptions))
	}
}
