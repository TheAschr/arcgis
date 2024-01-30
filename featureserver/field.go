package featureserver

const (
	// Equivalent to int32
	FieldTypeOID = "esriFieldTypeOID"
	// Equivalent to int16
	FieldTypeSmallInt = "esriFieldTypeSmallInteger"
	// Equivalent to int32
	FieldTypeInt = "esriFieldTypeInteger"
	// Equivalent to float32
	FieldTypeFloat = "esriFieldTypeFloat"
	// Equivalent to float64
	FieldTypeDouble = "esriFieldTypeDouble"
	FieldTypeString = "esriFieldTypeString"
	FieldTypeDate   = "esriFieldTypeDate"
)

type Field struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
	// Can be one of:
	//  - FieldTypeOID
	//  - FieldTypeSmallInt
	//  - FieldTypeInt
	//  - FieldTypeDouble
	//  - FieldTypeString
	//  - FieldTypeDate
	Type   string `json:"type"`
	Length int    `json:"length"`
}
