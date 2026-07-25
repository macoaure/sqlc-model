// Package plan builds the deterministic, order-preserved GenerationPlan
// (contexts -> models -> fields -> operations) from validated configuration,
// resolved field mappings, and contract-validated operations. Any
// error-severity diagnostic anywhere aborts plan construction entirely,
// which is the mechanism behind the generator's all-or-nothing guarantee.
package plan
