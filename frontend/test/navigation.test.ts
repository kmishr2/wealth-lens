import assert from "node:assert/strict";
import test from "node:test";
import { safeNextPath } from "../src/lib/navigation.ts";

test("safeNextPath accepts local absolute paths", () => {
  assert.equal(safeNextPath("/portfolios/123?tab=history"), "/portfolios/123?tab=history");
});

test("safeNextPath rejects external and protocol-relative redirects", () => {
  for (const value of [null, "", "https://evil.example", "//evil.example/path", "dashboard"]) {
    assert.equal(safeNextPath(value), "/dashboard");
  }
});
