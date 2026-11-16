package _07

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		wantErr error
	}{
		{name: "zero", percent: 0},
		{name: "regular discount", percent: 10},
		{name: "maximum discount", percent: 100},
		{name: "below minimum", percent: -1, wantErr: ErrInvalidPercent},
		{name: "above maximum", percent: 101, wantErr: ErrInvalidPercent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePercent(tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name    string
		price   int
		percent int
		want    int
		wantErr error
	}{
		{name: "regular discount", price: 1_000, percent: 15, want: 850},
		{name: "zero discount", price: 1_000, percent: 0, want: 1_000},
		{name: "full discount", price: 1_000, percent: 100, want: 0},
		{name: "negative price", price: -1, percent: 10, wantErr: ErrNegativePrice},
		{name: "negative percent", price: 1_000, percent: -1, wantErr: ErrInvalidPercent},
		{name: "percent above maximum", price: 1_000, percent: 101, wantErr: ErrInvalidPercent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Apply(tt.price, tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFinalPrice(t *testing.T) {
	tests := []struct {
		name    string
		items   []Item
		want    int
		wantErr error
	}{
		{name: "empty list", want: 0},
		{
			name: "multiple valid items",
			items: []Item{
				{Price: 1_000, DiscountPercent: 10},
				{Price: 500, DiscountPercent: 20},
				{Price: 200, DiscountPercent: 0},
			},
			want: 1_500,
		},
		{
			name:    "invalid item price",
			items:   []Item{{Price: -1, DiscountPercent: 10}},
			wantErr: ErrNegativePrice,
		},
		{
			name:    "invalid item discount",
			items:   []Item{{Price: 100, DiscountPercent: 101}},
			wantErr: ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FinalPrice(tt.items)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Zero(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
