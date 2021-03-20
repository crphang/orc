package orc

import (
	"bytes"
	"compress/flate"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func testWrite(writer *Writer) error {
	now := time.Unix(1478123411, 99).UTC()
	timeIncrease := 5*time.Second + 10001*time.Nanosecond
	length := 10001
	var intSum int64
	for i := 0; i < length; i++ {
		string1 := fmt.Sprintf("%x", rand.Int63n(1000))
		timestamp1 := now.Add(time.Duration(i) * timeIncrease)
		int1 := rand.Int63n(10000)
		intSum += int1
		boolean1 := int1 > 4444
		double1 := rand.Float64()
		nested := []interface{}{
			rand.Float64(),
			[]interface{}{
				rand.Int63n(10000),
			},
		}
		err := writer.Write(string1, timestamp1, int1, boolean1, double1, nested)
		if err != nil {
			return err
		}
	}

	err := writer.Close()
	return err
}

// BenchmarkWrite/write-12         	      79	  14747973 ns/op	 4254266 B/op	  181089 allocs/op
func BenchmarkWrite(b *testing.B) {
	buf := &bytes.Buffer{}

	schema, _ := ParseSchema("struct<int1:int,int2:int,double1:double,string1:string>")

	type row []interface{}
	rows := []row{}
	for i:=0; i<15000;i++ {
		currentRow := row{}
		currentRow = append(currentRow, 1)
		currentRow = append(currentRow, i)
		currentRow = append(currentRow, float64(i))
		currentRow = append(currentRow, fmt.Sprintf("%d", 1))

		rows = append(rows, currentRow)
	}

	// Run the actual benchmark
	b.Run("write", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			w, _ := NewWriter(buf, SetSchema(schema))

			for _, v := range rows {
				err := w.Write(v...)
				if err != nil {
					b.Fatal(err)
				}
			}
			_ = w.Close()
			buf.Reset()
		}
	})
}

// BenchmarkColumnWrite/write-12         	      87	  13563963 ns/op	 3774514 B/op	  166097 allocs/op
func BenchmarkColumnWrite(b *testing.B) {
	buf := &bytes.Buffer{}

	schema, _ := ParseSchema("struct<int1:int,int2:int,double1:double,string1:string>")
	col0 := &Column{
		Data: []interface{}{},
	}
	col1 := &Column{
		Data: []interface{}{},
	}
	col2 := &Column{
		Data: []interface{}{},
	}
	col3 := &Column{
		Data: []interface{}{},
	}
	for i:=0; i<15000;i++ {
		col0.Data = append(col0.Data, 1)
		col1.Data = append(col1.Data, i)
		col2.Data = append(col2.Data, float64(i))
		col3.Data = append(col3.Data, fmt.Sprintf("%d", 1))
	}
	// Flushing first set of columns
	iteratable := []ColumnIterator{col0, col1, col2, col3}

	b.Run("write", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			w, _ := NewWriter(buf, SetSchema(schema))
			err := w.WriteColumns(iteratable)
			if err != nil {
				b.Fatal(err)
			}
			_ = w.Close()
			buf.Reset()
		}
	})
}

// BenchmarkWriteSnappy/write-12         	      60	  19644749 ns/op	 6882465 B/op	  193156 allocs/op
func BenchmarkWriteSnappy(b *testing.B) {
	buf := &bytes.Buffer{}
	// Run the actual benchmark
	b.Run("write", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			schema, err := ParseSchema("struct<string1:string,timestamp1:timestamp,int1:int,boolean1:boolean,double1:double,nested:struct<double2:double,nested:struct<int2:int>>>")
			if err != nil {
				b.Fatal(err)
			}

			w, err := NewWriter(buf, SetSchema(schema), SetCompression(CompressionSnappy{}))
			if err != nil {
				b.Fatal(err)
			}

			_ = testWrite(w)
			buf.Reset()
		}
	})
}


// BenchmarkWriteZlib/write-12         	      39	  29914068 ns/op	27983502 B/op	  192054 allocs/op
func BenchmarkWriteZlib(b *testing.B) {
	buf := &bytes.Buffer{}
	// Run the actual benchmark
	b.Run("write", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for n := 0; n < b.N; n++ {
			schema, err := ParseSchema("struct<string1:string,timestamp1:timestamp,int1:int,boolean1:boolean,double1:double,nested:struct<double2:double,nested:struct<int2:int>>>")
			if err != nil {
				b.Fatal(err)
			}

			w, err := NewWriter(buf, SetSchema(schema), SetCompression(CompressionZlib{Level: flate.DefaultCompression}))
			if err != nil {
				b.Fatal(err)
			}

			_ = testWrite(w)

			buf.Reset()
		}
	})
}