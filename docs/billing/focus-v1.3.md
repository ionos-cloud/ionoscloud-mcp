# FOCUS v1.3 — Billing Output Reference

Assistant-formatted billing output, and any post-processed normalized output derived from billing tool responses, MUST use FOCUS v1.3 column names and value constraints. Raw billing tool responses may continue to use the Billing API's JSON field names.
Spec: https://focus.finops.org/focus-specification/v1-3/

## Columns (Cost & Usage Dataset)

M=Mandatory, C=Conditional, R=Recommended, O=Optional. Null means omit or set null.

### Account
| Column | Level | Type | Rule |
|---|---|---|---|
| BillingAccountId | M | String | IONOS contract number |
| BillingAccountName | M | String | Contract label |
| BillingAccountType | R | String | e.g. "Contract" |
| SubAccountId | M | String | IONOS customer ID |
| SubAccountName | M | String | Customer label |
| SubAccountType | R | String | e.g. "Customer" |

### Billing & Cost
| Column | Level | Type | Rule |
|---|---|---|---|
| BilledCost | M | Decimal | Invoice/charge amount |
| BillingCurrency | M | String | ISO 4217 (e.g. "EUR") |
| EffectiveCost | M | Decimal | Amortized cost after discounts; equals BilledCost when no commitments apply |
| ContractedCost | C | Decimal | Cost at negotiated rate. Null when no contract pricing. |
| ContractedUnitPrice | C | Decimal | Null when no contract pricing |
| ListCost | C | Decimal | Cost at on-demand/list rate |
| ListUnitPrice | C | Decimal | Null if no published price list |
| ConsumedQuantity | C | Decimal | Not null when ChargeCategory="Usage" and ChargeClass≠"Correction" |
| ConsumedUnit | C | String | FOCUS unit format (see below). Same nullability as ConsumedQuantity |

### Charge
| Column | Level | Type | Allowed Values |
|---|---|---|---|
| ChargeCategory | M | String | `Usage` · `Purchase` · `Tax` · `Credit` · `Adjustment` |
| ChargeClass | M | String | Null for normal charges · `Correction` for corrections/refunds/rebates |
| ChargeDescription | M | String | Human-readable line item description |
| ChargeFrequency | M | String | `One-Time` · `Recurring` · `Usage-Based` |
| ChargePeriodStart | M | DateTime | ISO 8601 / RFC 3339, inclusive |
| ChargePeriodEnd | M | DateTime | ISO 8601 / RFC 3339, exclusive |

### Timeframe
| Column | Level | Type |
|---|---|---|
| BillingPeriodStart | M | DateTime |
| BillingPeriodEnd | M | DateTime |

### Charge Origination
| Column | Level | Type | Rule |
|---|---|---|---|
| InvoiceIssuerName | M | String | "IONOS CLOUD" |
| ServiceProviderName | M | String | "IONOS CLOUD" |
| HostProviderName | R | String | "IONOS CLOUD" |
| InvoiceId | R | String | IONOS invoice ID when available |

### Location
| Column | Level | Type | Rule |
|---|---|---|---|
| AvailabilityZone | O | String | Null if not applicable |
| RegionId | R | String | IONOS datacenter UUID |
| RegionName | R | String | Datacenter display name (e.g. "Frankfurt") |

### Pricing
| Column | Level | Type | Rule |
|---|---|---|---|
| PricingCategory | C | String | `Standard` · `Committed` · `Interruptible` · `Dynamic` · `Other`. Not null for Usage/Purchase rows |
| PricingCurrency | C | String | ISO 4217. Required when differs from BillingCurrency |
| PricingQuantity | C | Decimal | Quantity in PricingUnit |
| PricingUnit | C | String | Provider pricing unit |
| PricingCurrencyContractedUnitPrice | C | Decimal | Null when no contract |
| PricingCurrencyEffectiveCost | C | Decimal | |
| PricingCurrencyListUnitPrice | C | Decimal | |

### Resource
| Column | Level | Type | Rule |
|---|---|---|---|
| ResourceId | R | String | IONOS resource UUID |
| ResourceName | R | String | Resource label |
| ResourceType | R | String | e.g. "Server", "Volume", "NIC" |
| Tags | R | JSON | Key-value pairs per FOCUS key-value format |

### Service
| Column | Level | Type | Rule |
|---|---|---|---|
| ServiceCategory | M | String | See allowed values below |
| ServiceName | M | String | IONOS service name (e.g. "Compute Engine", "Block Storage") |
| ServiceSubcategory | R | String | See FOCUS spec for allowed parent-child pairs |

ServiceCategory allowed: `AI and Machine Learning` · `Analytics` · `Business Application` · `Compute` · `Databases` · `Developer Tools` · `Identity` · `Integration` · `IoT` · `Management and Governance` · `Media` · `Migration` · `Multicloud` · `Networking` · `Security` · `Storage` · `Web` · `Other`

### SKU
| Column | Level | Type | Rule |
|---|---|---|---|
| SkuId | C | String | Required when provider publishes price lists |
| SkuMeter | C | JSON | FOCUS-specified property names + units |
| SkuPriceDetails | C | JSON | |
| SkuPriceId | C | String | |

### Commitment Discount (all Conditional — include only when applicable)
| Column | Type | Rule |
|---|---|---|
| CommitmentDiscountId | String | Unique ID of the commitment |
| CommitmentDiscountName | String | Display name |
| CommitmentDiscountCategory | String | `Spend` or `Usage` |
| CommitmentDiscountType | String | Provider-specific (e.g. "Reserved Instance") |
| CommitmentDiscountStatus | String | `Used` or `Unused` |
| CommitmentDiscountQuantity | Decimal | |
| CommitmentDiscountUnit | String | |

### Capacity Reservation (Conditional)
| Column | Type | Rule |
|---|---|---|
| CapacityReservationId | String | |
| CapacityReservationStatus | String | `Used` or `Unused` |

### Contract (Conditional)
| Column | Type | Rule |
|---|---|---|
| ContractApplied | JSON | FOCUS JSON object format with contract+commitment IDs |

### Allocation (Optional — v1.3 new)
| Column | Type |
|---|---|
| AllocatedMethodId | String |
| AllocatedMethodDetails | String |
| AllocatedResourceId | String |
| AllocatedResourceName | String |
| AllocatedTags | JSON |

## IONOS Tool → FOCUS Mapping

| Tool | ChargeCategory | ChargeFrequency | Key Mappings |
|---|---|---|---|
| list_billing_invoices | Per line type: fee→Usage, tax→Tax, credit→Credit, refund→Adjustment | Recurring or One-Time | id→InvoiceId, amount→BilledCost, unit→BillingCurrency, date→BillingPeriodStart |
| get_billing_invoice | Per line type (same as above) | Recurring or One-Time | metadata.invoiceId→InvoiceId, metadata.startDate→BillingPeriodStart, metadata.endDate→BillingPeriodEnd; per datacenter: id→ResourceId, name→ResourceName, location→RegionId; per meter: amount→BilledCost, rate→ListUnitPrice, quantity→ConsumedQuantity |
| list_billing_usage / get_billing_usage_by_datacenter | Usage | Usage-Based | CPU-hours/GB-hours→ConsumedQuantity+ConsumedUnit, id→ResourceId, name→ResourceName, location→RegionId |
| list_billing_traffic / list_billing_traffic_by_period | Usage | Usage-Based | In/Out→ConsumedQuantity, "Byte"→ConsumedUnit, vdcUUID→ResourceId, vdcName→ResourceName, ip→ResourceId (NIC rows) |
| list_billing_utilization / list_billing_utilization_by_period / get_billing_utilization_daily | Usage | Usage-Based | ConsumedQuantity+ConsumedUnit from quantity; resourceId→ResourceId, name→ResourceName, region→RegionId (meter-level); DC has id+name but no location field |
| list_billing_evn / list_billing_evn_by_period | Usage | Recurring | from/to→ChargePeriodStart/End, resourceUUID→ResourceId; vdcUUID→datacenter identifier (no location field — cannot map to RegionId) |
| list_billing_products | — | — | Reference only for SkuId, ListUnitPrice, PricingUnit lookups |
| get_billing_profile | — | — | contractId→BillingAccountId, customerId→SubAccountId |

## Format Rules

- **Dates**: ISO 8601 / RFC 3339 (e.g. `2025-03-01T00:00:00Z`)
- **Currency**: ISO 4217 three-letter code
- **Decimals**: No rounding rules specified; preserve source precision
- **Nulls**: Use null (not empty string) for inapplicable columns
- **Tags/JSON columns**: FOCUS key-value format `{"key":"value"}` — keys are strings, values are strings
- **Custom columns**: Prefix with `x_` (e.g. `x_IonosContractType`)
- **Unit format**: Compound units use dash separator (e.g. `Gb-hour`, `vCPU-hour`)
- **Column names**: PascalCase, max 50 chars
