// Package authservice re-exports auth service types for monolith use.
package authservice

import (
	cfg "github.com/FibrinLab/open-nucleus/services/auth/internal/config"
	svc "github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	s "github.com/FibrinLab/open-nucleus/services/auth/internal/store"
)

type AuthService = svc.AuthService
type SmartService = svc.SmartService

var NewAuthService = svc.NewAuthService
var NewSmartService = svc.NewSmartService

// Config re-exports the auth service's internal config for monolith construction.
type Config = cfg.Config
type JWTConfig = cfg.JWTConfig
type GitConfig = cfg.GitConfig
type NodeConfig = cfg.NodeConfig
type DevicesConfig = cfg.DevicesConfig
type SecurityConfig = cfg.SecurityConfig
type KeyStoreConfig = cfg.KeyStoreConfig
type SQLiteConfig = cfg.SQLiteConfig

type DenyList = s.DenyList
type ClientStore = s.ClientStore

var NewDenyList = s.NewDenyList
var NewClientStore = s.NewClientStore
var InitSchema = s.InitSchema
var InitClientSchema = s.InitClientSchema
