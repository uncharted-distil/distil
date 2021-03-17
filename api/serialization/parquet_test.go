package serialization

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteParquet(t *testing.T) {
	p := NewParquet()

	// Generate a parquet file
	rows := 5
	cols := 10
	fileName := "test/test_file.parquet"
	originalData, err := generateParquet(p, cols, rows, fileName)
	assert.NoError(t, err)

	// Read it in with a transposition
	data, err := p.ReadData(fileName)
	assert.NoError(t, err)
	assert.Equal(t, len(data), rows)
	assert.Equal(t, len(data[0]), cols)

	assert.Equal(t, originalData, data)
}

func generateParquet(p *Parquet, cols int, rows int, fileName string) ([][]string, error) {
	data := make([][]string, rows)
	for i := 0; i < rows; i++ {
		rowData := make([]string, cols)
		for j := 0; j < cols; j++ {
			rowData[j] = fmt.Sprintf("row_%d__col_%d", i, j)
		}
		data[i] = rowData
	}
	err := p.WriteData(fileName, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
