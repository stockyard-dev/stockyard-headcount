package server
type Limits struct{Tier string;Description string;MaxEmployees int}
func LimitsFor(tier string)Limits{if tier=="pro"{return Limits{Tier:"pro",Description:"Pro tier",MaxEmployees:0}};return Limits{Tier:"free",Description:"Free tier",MaxEmployees:50}}
func(l Limits)IsPro()bool{return l.Tier=="pro"}
