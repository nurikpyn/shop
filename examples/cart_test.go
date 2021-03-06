package examples_test

import (
	"fmt"

	"github.com/foomo/shop/examples"
	"github.com/foomo/shop/order"
)

const MOCK_PRICE = 9.99

func ExampleOrderCustom_createCart() {
	// a cart is an incomplete order
	o, _ := order.NewOrder(&examples.SmurfOrderCustomProvider{})
	o.Custom = &examples.SmurfOrderCustom{
		ResponsibleSmurf: "Pete",
	}
	const (
		positionIDA = "awesome-computer-a"
		positionIDB = "awesome-computer-b"
	)

	// add a product
	o.AddPosition(&order.Position{
		ItemID:   positionIDA,
		Name:     "an awesome computer",
		Quantity: 1.0,
		Custom: &examples.SmurfPositionCustom{
			Foo: "foo",
		},
	})

	// set qty
	if o.SetPositionQuantity(positionIDA, 3.01, MOCK_PRICE, MOCK_PRICE, nil) != nil {
		panic("could not set qty")
	}

	// add another position
	o.AddPosition(&order.Position{
		ItemID:   positionIDB,
		Name:     "an awesome computer",
		Quantity: 1.0,
		Custom: &examples.SmurfPositionCustom{
			Foo: "bar",
		},
	})

	o.SetPositionQuantity(positionIDB, 0, MOCK_PRICE, MOCK_PRICE, nil)

	fmt.Println(
		"responsible smurf:",
		o.Custom.(*examples.SmurfOrderCustom).ResponsibleSmurf,
		", position foo:",
		o.Positions[0].Custom.(*examples.SmurfPositionCustom).Foo,
		", qty:",
		o.Positions[0].Quantity,
		", number of positions:",
		len(o.Positions),
	)
	// Output: responsible smurf: Pete , position foo: foo , qty: 3.01 , number of positions: 1
}
