package nfz

import nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"

// Specialties is the curated list behind the start screen's tiles. Every Benefit/BenefitChild is an
// exact name from NFZ's /benefits dictionary (verified against the live API) -- /queues needs exact names.
var Specialties = []nfzpg.Specialty{
	{Key: "alergolog", Label: "Alergolog", Icon: "allergy", Benefit: "PORADNIA ALERGOLOGICZNA", BenefitChild: "PORADNIA ALERGOLOGICZNA DLA DZIECI"},
	{Key: "laryngolog", Label: "Laryngolog", Icon: "ear", Benefit: "ŚWIADCZENIA Z ZAKRESU OTOLARYNGOLOGII"},
	{Key: "okulista", Label: "Okulista", Icon: "eye", Benefit: "ŚWIADCZENIA Z ZAKRESU OKULISTYKI"},
	{Key: "ortopeda", Label: "Ortopeda", Icon: "bone", Benefit: "ŚWIADCZENIA Z ZAKRESU ORTOPEDII I TRAUMATOLOGII NARZĄDU RUCHU"},
	{Key: "neurolog", Label: "Neurolog", Icon: "brain", Benefit: "ŚWIADCZENIA Z ZAKRESU NEUROLOGII"},
	{Key: "dermatolog", Label: "Dermatolog", Icon: "skin", Benefit: "PORADNIA DERMATOLOGICZNA", BenefitChild: "PORADNIA DERMATOLOGICZNA DLA DZIECI"},
	{Key: "logopeda", Label: "Logopeda", Icon: "speech", Benefit: "PORADNIA LOGOPEDYCZNA", BenefitChild: "PORADNIA LOGOPEDYCZNA DLA DZIECI"},
	{Key: "endokrynolog", Label: "Endokrynolog", Icon: "hormone", Benefit: "ŚWIADCZENIA Z ZAKRESU ENDOKRYNOLOGII"},
	{Key: "gastrolog", Label: "Gastrolog", Icon: "stomach", Benefit: "ŚWIADCZENIA Z ZAKRESU GASTROENTEROLOGII"},
	{Key: "psycholog", Label: "Psycholog", Icon: "mind", Benefit: "PORADNIA PSYCHOLOGICZNA", BenefitChild: "PORADNIA PSYCHOLOGICZNA DLA DZIECI"},
	{Key: "psychiatra", Label: "Zdrowie psychiczne", Icon: "heart", Benefit: "PORADNIA ZDROWIA PSYCHICZNEGO", BenefitChild: "PORADNIA ZDROWIA PSYCHICZNEGO DLA DZIECI"},
	{Key: "rehabilitacja", Label: "Rehabilitacja", Icon: "rehab", Benefit: "PORADNIA REHABILITACYJNA", BenefitChild: "PORADNIA REHABILITACYJNA DLA DZIECI"},
}
