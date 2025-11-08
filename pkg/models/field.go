package models

import (
	"encoding/json"
	"time"
)

type Field struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description *string          `json:"description"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
	Location    *json.RawMessage `json:"location"`
	Square      *int             `json:"square"`
	Info        *string          `json:"info"`
	Places      int              `json:"places"`
	Dressing    bool             `json:"dressing"`
	Toilet      bool             `json:"toilet"`
	Display     bool             `json:"display"`
	Parking     bool             `json:"parking"`
	ForDisabled bool             `json:"for_disabled"`
	Logo        *string          `json:"logo"`
	Media       *json.RawMessage `json:"media"`
	Status      *int             `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   *time.Time       `json:"deleted_at"`
}

type FieldView struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description *string          `json:"description"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
	Location    *json.RawMessage `json:"location" swaggertype:"string"`
	Square      *int             `json:"square"`
	Info        *string          `json:"info"`
	Places      int              `json:"places"`
	Dressing    bool             `json:"dressing"`
	Toilet      bool             `json:"toilet"`
	Display     bool             `json:"display"`
	Parking     bool             `json:"parking"`
	ForDisabled bool             `json:"for_disabled"`
	Logo        *Media           `json:"logo"`  // Логотип
	Media       *[]Media         `json:"media"` // Медиа
	Status      *int             `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
}

type CreateFieldRequest struct {
	Name        string           `json:"name" validate:"required"`
	Slug        string           `json:"slug"`
	Description *string          `json:"description"`
	City        string           `json:"city" validate:"required"`
	Address     string           `json:"address" validate:"required"`
	Location    *json.RawMessage `json:"location" swaggertype:"string"`
	Square      *int             `json:"square"`
	Info        *string          `json:"info"`
	Places      int              `json:"places"`
	Dressing    bool             `json:"dressing"`
	Toilet      bool             `json:"toilet"`
	Display     bool             `json:"display"`
	Parking     bool             `json:"parking"`
	ForDisabled bool             `json:"for_disabled"`
	Logo        *string          `json:"logo"`
	Media       *json.RawMessage `json:"media" swaggertype:"string"`
}

type UpdateFieldRequest struct {
	Name        string           `json:"name" validate:"required"`
	Slug        string           `json:"slug"`
	Description *string          `json:"description"`
	City        string           `json:"city" validate:"required"`
	Address     string           `json:"address" validate:"required"`
	Location    *json.RawMessage `json:"location" swaggertype:"string"`
	Square      *int             `json:"square"`
	Info        *string          `json:"info"`
	Places      int              `json:"places"`
	Dressing    bool             `json:"dressing"`
	Toilet      bool             `json:"toilet"`
	Display     bool             `json:"display"`
	Parking     bool             `json:"parking"`
	ForDisabled bool             `json:"for_disabled"`
	Logo        *string          `json:"logo"`
	Media       *json.RawMessage `json:"media" swaggertype:"string"`
}
