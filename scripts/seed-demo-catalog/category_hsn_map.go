package main

// CategoryMeta maps each target category to a valid HSN code and its returnability flag
type CategoryMeta struct {
	HSNCode      string
	IsReturnable bool
}

var CategoryHSNMap = map[string]CategoryMeta{
	"cereals":       {HSNCode: "1101", IsReturnable: false},
	"staples":       {HSNCode: "1101", IsReturnable: false},
	"dairy":         {HSNCode: "0401", IsReturnable: false},
	"snacks":        {HSNCode: "1905", IsReturnable: false},
	"beverages":     {HSNCode: "2202", IsReturnable: false},
	"personal-care": {HSNCode: "3401", IsReturnable: true},
	"household":     {HSNCode: "3402", IsReturnable: true},
	"biscuits":      {HSNCode: "1905", IsReturnable: false},
	"chocolates":    {HSNCode: "1806", IsReturnable: false},
	"frozen":        {HSNCode: "0405", IsReturnable: false},
}
