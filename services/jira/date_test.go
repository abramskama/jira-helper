package jira

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertDate(t *testing.T) {
	currentDate := time.Now()

	t.Run("today date", func(t *testing.T) {
		got, err := convertDate("today")
		require.NoError(t, err)

		assert.Equal(t, currentDate.Format(time.DateOnly), got)
	})

	t.Run("yesterday date", func(t *testing.T) {
		got, err := convertDate("yest")
		require.NoError(t, err)

		assert.Equal(t, currentDate.AddDate(0, 0, -1).Format(time.DateOnly), got)
	})

	tests := []struct {
		name    string
		date    string
		want    string
		wantErr error
	}{
		{
			name: "valid date with YYYY-MM-DD format",
			date: "2024-04-17",
			want: "2024-04-17",
		},
		{
			name:    "invalid format, got error",
			date:    "invalid",
			wantErr: ErrParseDatFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertDate(tt.date)

			require.ErrorIs(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}

}
