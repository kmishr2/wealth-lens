# Financial Engine

## Purpose

The financial engine contains all deterministic financial calculations used by the application.

This module must:

- remain framework-independent
- remain testable
- remain explainable
- avoid hidden logic

---

# Package Structure

```txt
/pkg/finance
    /allocation
    /cagr
    /xirr
    /volatility
    /drawdown
    /beta
    /rebalancing
    /taxation
    /healthscore
```

---

# Financial Metrics

## CAGR

Used for:

- long-term performance evaluation
- benchmark comparison

---

## XIRR

Used for:

- irregular cash flow return calculations
- SIP performance analysis

---

## Volatility

Used for:

- portfolio risk measurement
- benchmark-relative risk

---

## Maximum Drawdown

Used for:

- downside risk analysis
- historical decline analysis

---

## Beta

Used for:

- benchmark-relative movement analysis
- equity exposure estimation

---

## Allocation Drift

Used for:

- rebalancing analysis
- target allocation comparison

---

# Metric Definition Requirements

Every metric must define:

- name
- formula
- assumptions
- required inputs
- explanation text

---

# Testing Requirements

Every formula requires:

- unit tests
- edge-case tests
- deterministic outputs
- precision validation

---

# Numerical Accuracy

Requirements:

- avoid floating-point issues where possible
- consistent rounding rules
- reproducible calculations
- deterministic outputs

---

# Forbidden Logic

Do NOT include:

- predictions
- heuristics
- AI scoring
- probabilistic recommendations
- hidden adjustments

---
