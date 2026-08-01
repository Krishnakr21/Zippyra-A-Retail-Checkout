package main

type PriceRange struct {
	MinPaise int64
	MaxPaise int64
}

// CategoryPriceRanges covers realistic Indian retail price bands for each target category
var CategoryPriceRanges = map[string]PriceRange{
	"cereals":       {MinPaise: 15000, MaxPaise: 50000}, // ₹150 - ₹500
	"staples":       {MinPaise: 10000, MaxPaise: 45000}, // ₹100 - ₹450
	"dairy":         {MinPaise: 2500, MaxPaise: 12000},  // ₹25 - ₹120
	"snacks":        {MinPaise: 1000, MaxPaise: 15000},  // ₹10 - ₹150
	"beverages":     {MinPaise: 1500, MaxPaise: 20000},  // ₹15 - ₹200
	"personal-care": {MinPaise: 4000, MaxPaise: 35000},  // ₹40 - ₹350
	"household":     {MinPaise: 3000, MaxPaise: 40000},  // ₹30 - ₹400
	"biscuits":      {MinPaise: 2000, MaxPaise: 18000},  // ₹20 - ₹180
	"chocolates":    {MinPaise: 1500, MaxPaise: 25000},  // ₹15 - ₹250
	"frozen":        {MinPaise: 8000, MaxPaise: 35000},  // ₹80 - ₹350
}
