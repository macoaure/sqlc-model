// Package codegen renders every generator-owned file kind (session,
// field-identifier constants, model, collection, internal store/record
// adapters) plus the one-time developer-owned extension file, from a
// GenerationPlan. Every rendered file is formatted and import-fixed before
// being placed in the plugin's response.
package codegen
