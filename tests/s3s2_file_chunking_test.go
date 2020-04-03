package main_test

import (
	"testing"
	file "github.com/tempuslabs/s3s2/file"
	"github.com/stretchr/testify/assert"
)




func TestChunkSplitEven(t * testing.T) {
    assert := assert.New(t)

    array := []file.File{
    file.File{"testfile0"},
    file.File{"testfile1"},
    file.File{"testfile2"},
    file.File{"testfile3"},
    file.File{"testfile4"},
    file.File{"testfile5"},
    file.File{"testfile6"},
    file.File{"testfile7"},
    }

    expected := [][]file.File{
    []file.File{file.File{"testfile0"}, file.File{"testfile1"}},
    []file.File{file.File{"testfile2"}, file.File{"testfile3"}},
    []file.File{file.File{"testfile4"}, file.File{"testfile5"}},
    []file.File{file.File{"testfile6"}, file.File{"testfile7"}},
    }

    actual := file.ChunkArray(array, 2)

    assert.Equal(actual, expected)
}

func TestChunkSplitOdd(t * testing.T) {
    assert := assert.New(t)

    array := []file.File{
    file.File{"testfile0"},
    file.File{"testfile1"},
    file.File{"testfile2"},
    file.File{"testfile3"},
    file.File{"testfile4"},
    file.File{"testfile5"},
    file.File{"testfile6"},
    file.File{"testfile7"},
    file.File{"testfile8"},
    file.File{"testfile9"},
    }

    expected := [][]file.File{
    []file.File{file.File{"testfile0"},file.File{"testfile1"},file.File{"testfile2"}},
    []file.File{file.File{"testfile3"},file.File{"testfile4"},file.File{"testfile5"}},
    []file.File{file.File{"testfile6"},file.File{"testfile7"},file.File{"testfile8"}},
    []file.File{file.File{"testfile9"}},
    }

    actual := file.ChunkArray(array, 3)

    assert.Equal(actual, expected)
}


func TestFileChunking(t *testing.T) {
    TestChunkSplitEven(t)
    TestChunkSplitOdd(t)
}