package testing

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCheckListResourceAttr(name, key string, values ...string) resource.TestCheckFunc {
	valueCheckFuncs := make([]resource.TestCheckFunc, len(values)+1)
	valueCheckFuncs[0] = resource.TestCheckResourceAttr(name, fmt.Sprintf("%s.#", key), fmt.Sprintf("%d", len(values)))
	for i, value := range values {
		valueCheckFuncs[i+1] = resource.TestCheckResourceAttr(name, fmt.Sprintf("%s.%d", key, i), value)
	}
	return resource.ComposeTestCheckFunc(valueCheckFuncs...)
}
