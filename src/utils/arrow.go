package utils

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
)

// PGTypeToArrowType maps PostgreSQL type OIDs to Arrow data types.
// It returns the corresponding Arrow data type for a given PostgreSQL type OID.
// If the PostgreSQL type OID is not supported, it returns an error.
func PGTypeToArrowType(pgTypeOID uint32) (arrow.DataType, error) {
	switch pgTypeOID {
	case 20: // BIGINT
		return arrow.PrimitiveTypes.Int64, nil
	case 21: // SMALLINT
		return arrow.PrimitiveTypes.Int16, nil
	case 23: // INTEGER
		return arrow.PrimitiveTypes.Int32, nil
	case 700: // REAL
		return arrow.PrimitiveTypes.Float32, nil
	case 701: // DOUBLE PRECISION
		return arrow.PrimitiveTypes.Float64, nil
	case 1042, 1043: // CHAR, VARCHAR
		return arrow.BinaryTypes.String, nil
	case 1700: // NUMERIC
		return arrow.PrimitiveTypes.Float64, nil
	case 16: // BOOLEAN
		return arrow.FixedWidthTypes.Boolean, nil
	case 25: // TEXT
		return arrow.BinaryTypes.String, nil
	case 1114, 1184: // TIMESTAMP WITHOUT TIME ZONE, TIMESTAMP WITH TIME ZONE
		return arrow.FixedWidthTypes.Timestamp_ms, nil
	case 1082: // DATE
		return arrow.FixedWidthTypes.Date32, nil
	case 114, 3802: // JSON, JSONB
		return arrow.BinaryTypes.String, nil
	// Add more type mappings as needed
	default:
		return nil, fmt.Errorf("unsupported PostgreSQL type OID: %d", pgTypeOID)
	}
}
