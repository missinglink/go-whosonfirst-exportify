package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer"
	"github.com/whosonfirst/go-writer"
	"log"
	"os"
	"strconv"
)

func main() {

	reader_uri := flag.String("reader-uri", "", "A valid whosonfirst/go-reader URI.")
	writer_uri := flag.String("writer-uri", "", "A valid whosonfirst/go-writer URI. If empty the value of the -reader-uri flag will be assumed.")
	parent_reader_uri := flag.String("parent-reader-uri", "", "A valid whosonfirst/go-reader URI. If empty the value of the -reader-uri flag will be assumed.")

	exporter_uri := flag.String("exporter-uri", "whosonfirst://", "A valid whosonfirst/go-whosonfirst-export URI.")

	parent_id := flag.Int64("parent-id", 0, "A valid Who's On First ID.")

	from_stdin := flag.Bool("stdin", false, "Read target IDs from STDIN")
	
	flag.Usage = func() {

		fmt.Fprintf(os.Stderr, "Assign the parent ID and its hierarchy to one or more WOF records\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] target-id-(N) target-id-(N)\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "For example:\n")
		// fmt.Fprintf(os.Stderr, "\t%s -reader-uri fs:///usr/local/data/sfomuseum-data-architecture/data -parent-id 1477855937 1477855955\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// START OF put me in a function...

	str_ids := flag.Args()

	if *from_stdin {

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			str_ids = append(str_ids, scanner.Text())
		}

		err := scanner.Err()

		if err != nil {
			log.Fatalf("Failed to read input from STDIN, %v", err)
		}
	}

	target_ids := make([]int64, len(str_ids))

	for idx, this_id := range str_ids {

		id, err := strconv.ParseInt(this_id, 10, 64)

		if err != nil {
			log.Fatalf("Failed to parse '%s', %v", this_id, err)
		}

		target_ids[idx] = id
	}

	// END OF put me in a function
	
	if *parent_reader_uri == "" {
		*parent_reader_uri = *reader_uri
	}

	if *writer_uri == "" {
		*writer_uri = *reader_uri
	}

	ctx := context.Background()

	r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader for '%s', %v", *reader_uri, err)
	}

	parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader for '%s', %v", *parent_reader_uri, err)
	}

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer for '%s', %v", *writer_uri, err)
	}

	ex, err := export.NewExporter(ctx, *exporter_uri)

	if err != nil {
		log.Fatalf("Failed to create new exporter for '%s', %v", *exporter_uri, err)
	}

	// Parent stuff we only need to set up once

	parent_f, err := wof_reader.LoadBytesFromID(ctx, parent_r, *parent_id)

	if err != nil {
		log.Fatalf("Failed to load '%d', %v", *parent_id, err)
	}

	hier_rsp := gjson.GetBytes(parent_f, "properties.wof:hierarchy")

	if !hier_rsp.Exists() {
		log.Fatalf("Parent (%d) is missing properties.wof:hierarchy", *parent_id)
	}

	parent_hierarchy := hier_rsp.Value()

	to_update := map[string]interface{}{
		"properties.wof:parent_id": *parent_id,
		"properties.wof:hierarchy": parent_hierarchy,
	}

	// Okay, go

	for _, id := range target_ids {

		f, err := wof_reader.LoadBytesFromID(ctx, r, id)

		if err != nil {
			log.Fatalf("Failed to load '%d', %v", id, err)
		}

		for path, v := range to_update {

			f, err = sjson.SetBytes(f, path, v)

			if err != nil {
				log.Fatalf("Failed to update '%s', %v", path, err)
			}
		}

		f, err = ex.Export(ctx, f)

		if err != nil {
			log.Fatalf("Failed to export '%d', %v", id, err)
		}

		err = wof_writer.WriteFeatureBytes(ctx, wr, f)

		if err != nil {
			log.Fatalf("Failed to write '%d', %v", id, err)
		}
	}

}
