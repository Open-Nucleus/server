// Package formularyservice re-exports formulary service types for monolith use.
package formularyservice

import (
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/dosing"
	svc "github.com/FibrinLab/open-nucleus/services/formulary/internal/service"
	s "github.com/FibrinLab/open-nucleus/services/formulary/internal/store"
)

// DosingEngine re-exports the dosing engine interface and stub.
type DosingEngine = dosing.Engine
type StubDosingEngine = dosing.StubEngine

var NewStubDosingEngine = dosing.NewStubEngine
var NewPharmDosingEngine = dosing.NewPharmEngine

type FormularyService = svc.FormularyService

var New = svc.New

type DrugDB = s.DrugDB
type InteractionIndex = s.InteractionIndex
type StockStore = s.StockStore
type MedicationRecord = s.MedicationRecord
type AllergyMatch = s.AllergyMatch

type SearchResult = svc.SearchResult
type InteractionCheckResult = svc.InteractionCheckResult
type AllergyConflictResult = svc.AllergyConflictResult
type DeliveryItemInput = svc.DeliveryItemInput
type StockPrediction = svc.StockPrediction
type RedistributionSuggestion = svc.RedistributionSuggestion
type FormularyInfo = svc.FormularyInfo
type DosingWarningItem = svc.DosingWarningItem
type StockCheckItem = svc.StockCheckItem
type StockLevel = s.StockLevel
type InteractionRule = s.InteractionRule

var NewDrugDB = s.NewDrugDB
var NewInteractionIndex = s.NewInteractionIndex
var NewStockStore = s.NewStockStore
var InitSchema = s.InitSchema
