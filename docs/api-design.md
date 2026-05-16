# API Design

## API Style

- REST-first
- JSON-based
- stateless
- versioned endpoints

Base path:

```txt
/api/v1
```

---

# Core Resources

## Users

```txt
/users
```

---

## Portfolios

```txt
/portfolios
```

---

## Transactions

```txt
/transactions
```

---

## Holdings

```txt
/holdings
```

---

## Analytics

```txt
/analytics
```

---

## Benchmarks

```txt
/benchmarks
```

---

## Goals

```txt
/goals
```

---

# Response Standards

Success responses:

```json
{
  "success": true,
  "data": {}
}
```

Error responses:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid transaction"
  }
}
```

---

# Validation Rules

All write operations require:

- schema validation
- ownership validation
- numerical validation
- asset type validation

---

# Security Requirements

- JWT authentication
- password hashing
- rate limiting
- ownership-based authorization
- audit-safe transaction handling

---

# Performance Goals

- pagination for list endpoints
- snapshot-backed analytics
- avoid expensive historical recalculations

---
