// Package pipeline re-exports the write pipeline for use by the monolith.
package pipeline

import p "github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"

type Writer = p.Writer
type WriteResult = p.WriteResult
type BatchResult = p.BatchResult
type BatchItem = p.BatchItem
type BatchItemResult = p.BatchItemResult
type MutationContext = p.MutationContext
type ValidationError = p.ValidationError

var NewWriter = p.NewWriter
