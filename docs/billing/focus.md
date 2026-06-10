---
subcategory: "Billing"
page_title: "FOCUS Specification"
description: |-
  Tool for retrieving the FOCUS v1.3 billing specification and IONOS field mappings.
---

# FOCUS Specification

## get_billing_focus_spec

Returns the [FOCUS v1.3](https://focus.finops.org/) column specification together with the mappings from IONOS billing tool outputs (invoices, usage, traffic, utilization) to FOCUS fields. Call this before converting IONOS billing data to FOCUS format, or whenever FOCUS-compliant cost output is requested.

The same document is also exposed as the MCP resource `ionos://billing/focus-v1.3` for clients that support `resources/read`. This tool exists so that clients without resource support can still retrieve the specification.

**Parameters:** None

**Example:**

```json
{
  "name": "get_billing_focus_spec",
  "arguments": {}
}
```

**Reference:** [FOCUS — FinOps Open Cost and Usage Specification](https://focus.finops.org/)
