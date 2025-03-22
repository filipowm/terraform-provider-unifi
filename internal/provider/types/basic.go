package types

import "github.com/hashicorp/terraform-plugin-framework/types"

func StringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func Int32OrNull(i int) types.Int32 {
	if i == 0 {
		return types.Int32Null()
	}
	return types.Int32Value(int32(i))
}

func Int64OrNull(i int) types.Int64 {
	if i == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(i))
}
