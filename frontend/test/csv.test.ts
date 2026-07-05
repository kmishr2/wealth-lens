import assert from "node:assert/strict";
import test from "node:test";
import { csvCell, safeFilename, safeSpreadsheetText } from "../src/lib/csv.ts";

test("csvCell quotes values and doubles embedded quotes", () => {
  assert.equal(csvCell('Bank "Primary", INR'), '"Bank ""Primary"", INR"');
});

test("safeSpreadsheetText neutralizes spreadsheet formulas", () => {
  for (const value of ["=1+1", "+cmd", "-2+3", "@SUM(A1:A2)"]) {
    assert.equal(safeSpreadsheetText(value), `'${value}`);
  }
  assert.equal(safeSpreadsheetText("ordinary description"), "ordinary description");
});

test("safeFilename produces an ASCII attachment name with fallback", () => {
  assert.equal(safeFilename("Primary INR Portfolio"), "primary-inr-portfolio");
  assert.equal(safeFilename("  Wealth / 2026  "), "wealth-2026");
  assert.equal(safeFilename("₹₹₹"), "portfolio");
});
