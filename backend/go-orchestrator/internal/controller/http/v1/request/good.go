package request

type CreateGoodRequest struct {
	Name     string  `json:"name" binding:"required"`
	Weight   float64 `json:"weight" binding:"required,gt=0"`
	Height   float64 `json:"height" binding:"required,gt=0"`
	Length   float64 `json:"length" binding:"required,gt=0"`
	Width    float64 `json:"width" binding:"required,gt=0"`
	Quantity int     `json:"quantity" binding:"required,gt=0"`
}

type UpdateGoodRequest struct {
	Name   string  `json:"name" binding:"required"`
	Weight float64 `json:"weight" binding:"required,gt=0"`
	Height float64 `json:"height" binding:"required,gt=0"`
	Length float64 `json:"length" binding:"required,gt=0"`
	Width  float64 `json:"width" binding:"required,gt=0"`
}
