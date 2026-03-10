// Package anchorservice re-exports anchor service types for monolith use.
package anchorservice

import (
	svc "github.com/FibrinLab/open-nucleus/services/anchor/internal/service"
	s "github.com/FibrinLab/open-nucleus/services/anchor/internal/store"
)

type AnchorService = svc.AnchorService

var New = svc.New

type AnchorStore = s.AnchorStore
type CredentialStore = s.CredentialStore
type DIDStore = s.DIDStore
type AnchorQueue = s.AnchorQueue
type AnchorRecordData = s.AnchorRecordData
type QueueEntry = s.QueueEntry

type StatusResult = svc.StatusResult
type TriggerResult = svc.TriggerResult
type VerifyResult = svc.VerifyResult
type BackendInfo = svc.BackendInfo
type BackendStatusResult = svc.BackendStatusResult
type QueueStatusResult = svc.QueueStatusResult

var NewAnchorStore = s.NewAnchorStore
var NewCredentialStore = s.NewCredentialStore
var NewDIDStore = s.NewDIDStore
var NewAnchorQueue = s.NewAnchorQueue
var InitSchema = s.InitSchema
