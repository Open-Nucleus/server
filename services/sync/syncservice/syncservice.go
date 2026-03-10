// Package syncservice re-exports sync service types for monolith use.
package syncservice

import (
	cfg "github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	svc "github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	s "github.com/FibrinLab/open-nucleus/services/sync/internal/store"
)

type SyncEngine = svc.SyncEngine
type EventBus = svc.EventBus

var NewSyncEngine = svc.NewSyncEngine
var NewEventBus = svc.NewEventBus

type Event = svc.Event

const EventConflictResolved = svc.EventConflictResolved

// Config re-exports the sync service's internal config for monolith construction.
type Config = cfg.Config
type GitConfig = cfg.GitConfig

type ConflictStore = s.ConflictStore
type ConflictRecord = s.ConflictRecord
type HistoryStore = s.HistoryStore
type HistoryRecord = s.HistoryRecord
type PeerStore = s.PeerStore
type PeerRecord = s.PeerRecord

var NewConflictStore = s.NewConflictStore
var NewHistoryStore = s.NewHistoryStore
var NewPeerStore = s.NewPeerStore
var InitSchema = s.InitSchema
