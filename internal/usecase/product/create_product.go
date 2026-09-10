package product

import (
	"context"
	"time"

	"github.com/ViniciusVaz/gobid/internal/validator"
	"github.com/google/uuid"
)

type CreateProductRequest struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	Baseprice   float64   `json:"baseprice"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionDuration = 2 * time.Hour

func (req CreateProductRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "this field cannot be blank")
	eval.CheckField(validator.NotBlank(req.Description), "description", "this field cannot be blank")
	eval.CheckField(
		validator.MinChars(req.Description, 10) && validator.MaxChars(req.Description, 255),
		"description",
		"this field must have a length between 10 and 255",
	)
	eval.CheckField(req.Baseprice > 0, "baseprice", "this field must be greater or equal to 0")
	eval.CheckField(req.AuctionEnd.Sub(time.Now()) >= minAuctionDuration, "auction_end", "must be at least two hours duration")

	return eval
}
